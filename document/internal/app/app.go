package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	api "github.com/ossirdian/magistral/document/internal/api/documents/v1"
	studentclient "github.com/ossirdian/magistral/document/internal/client/student"
	"github.com/ossirdian/magistral/document/internal/config/env"
	"github.com/ossirdian/magistral/document/internal/httpauth"
	"github.com/ossirdian/magistral/document/internal/migrator"
	documentrepo "github.com/ossirdian/magistral/document/internal/repository/document"
	documentservice "github.com/ossirdian/magistral/document/internal/service/document"
	localstorage "github.com/ossirdian/magistral/document/internal/storage/local"
	auditpkg "github.com/ossirdian/magistral/platform/pkg/audit"
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
	closer.AddNamed("document-tracer", ptracing.ShutdownTracer)

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
	closer.AddNamed("document-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	storage := localstorage.New(cfg.Storage().Root())
	if storage == nil {
		return nil, errors.New("storage root is invalid")
	}

	studentConn, err := grpc.NewClient(
		cfg.Student().Addr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(ptracing.UnaryClientInterceptor("document")),
	)
	if err != nil {
		return nil, err
	}
	closer.AddNamed("document-student-grpc-client", func(context.Context) error { return studentConn.Close() })
	studentValidator := studentclient.New(studentv1.NewStudentServiceClient(studentConn))

	handler, err := api.NewServer(
		documentservice.NewService(
			documentrepo.NewRepository(pool),
			storage,
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
		Handler:           ptracing.HTTPMiddleware("document-http", pmetrics.HTTPMiddleware(auditpkg.Middleware("document", httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins())))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("document-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting document http server", "addr", a.server.Addr)
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
