package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	api "github.com/ossirdian/magistral/lesson/internal/api/lessons/v1"
	"github.com/ossirdian/magistral/lesson/internal/config/env"
	"github.com/ossirdian/magistral/lesson/internal/httpauth"
	"github.com/ossirdian/magistral/lesson/internal/migrator"
	"github.com/ossirdian/magistral/lesson/internal/planning"
	lessonrepo "github.com/ossirdian/magistral/lesson/internal/repository/lesson"
	lessonservice "github.com/ossirdian/magistral/lesson/internal/service/lesson"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
)

type App struct {
	server *http.Server
}

func New(ctx context.Context, cfg *env.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("lesson-tracer", ptracing.ShutdownTracer)

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
	closer.AddNamed("lesson-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	planningProducer, err := planning.NewProducer(cfg.Planning().KafkaBootstrapServers(), cfg.Planning().KafkaRequestTopic())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("lesson-planning-producer", func(context.Context) error {
		return planningProducer.Close()
	})

	handler, err := api.NewServer(lessonservice.NewService(lessonrepo.NewRepository(pool)), planningProducer)
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
		Handler:           ptracing.HTTPMiddleware("lesson-http", pmetrics.HTTPMiddleware(httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins()))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("lesson-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting lesson http server", "addr", a.server.Addr)
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
