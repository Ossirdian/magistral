package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	auditpkg "github.com/ossirdian/magistral/platform/pkg/audit"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	api "github.com/ossirdian/magistral/resources/internal/api/resources/v1"
	"github.com/ossirdian/magistral/resources/internal/config/env"
	"github.com/ossirdian/magistral/resources/internal/httpauth"
	"github.com/ossirdian/magistral/resources/internal/migrator"
	"github.com/ossirdian/magistral/resources/internal/repository/car"
	"github.com/ossirdian/magistral/resources/internal/repository/instructor"
	carservice "github.com/ossirdian/magistral/resources/internal/service/car"
	instructorservice "github.com/ossirdian/magistral/resources/internal/service/instructor"
)

type App struct {
	server *http.Server
}

func New(ctx context.Context, cfg *env.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("resources-tracer", ptracing.ShutdownTracer)

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
	closer.AddNamed("resources-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	handler, err := api.NewServer(
		carservice.NewService(car.NewRepository(pool)),
		instructorservice.NewService(instructor.NewRepository(pool)),
		cfg.Resources().SchoolID(),
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

	handlerWithMiddleware := ptracing.HTTPMiddleware(
		"resources-http",
		pmetrics.HTTPMiddleware(auditpkg.Middleware("resources", httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins()))),
	)

	server := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           handlerWithMiddleware,
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("resources-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting resources http server", "addr", a.server.Addr)
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
