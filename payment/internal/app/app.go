package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	api "github.com/ossirdian/magistral/payment/internal/api/payments/v1"
	studentclient "github.com/ossirdian/magistral/payment/internal/client/student"
	"github.com/ossirdian/magistral/payment/internal/config/env"
	"github.com/ossirdian/magistral/payment/internal/httpauth"
	"github.com/ossirdian/magistral/payment/internal/migrator"
	paymentrepo "github.com/ossirdian/magistral/payment/internal/repository/payment"
	paymentservice "github.com/ossirdian/magistral/payment/internal/service/payment"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	studentv1 "github.com/ossirdian/magistral/shared/pkg/student/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	server *http.Server
}

func New(ctx context.Context, cfg *env.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("payment-tracer", ptracing.ShutdownTracer)

	if cfg.Postgres().AutoMigrate() {
		db, err := sql.Open("pgx", cfg.Postgres().DSN())
		if err != nil {
			return nil, err
		}
		defer func() { _ = db.Close() }()

		if err := migrator.NewMigrator(db, cfg.Postgres().MigrationsDir()).Up(); err != nil {
			return nil, err
		}
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres().DSN())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("payment-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	studentConn, err := grpc.NewClient(
		cfg.Student().Addr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(ptracing.UnaryClientInterceptor("payment")),
	)
	if err != nil {
		return nil, err
	}
	closer.AddNamed("payment-student-grpc-client", func(context.Context) error { return studentConn.Close() })
	studentValidator := studentclient.New(studentv1.NewStudentServiceClient(studentConn))

	handler, err := api.NewServer(
		paymentservice.NewService(
			paymentrepo.NewRepository(pool),
			studentValidator,
		),
	)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/metrics", pmetrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           ptracing.HTTPMiddleware("payment-http", pmetrics.HTTPMiddleware(httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins()))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("payment-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting payment http server", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return closer.CloseAll(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
