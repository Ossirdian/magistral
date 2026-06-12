package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/audit/internal/config"
	"github.com/ossirdian/magistral/audit/internal/store"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
)

type App struct {
	server *http.Server
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg); err != nil {
		return nil, err
	}
	closer.AddNamed("audit-tracer", ptracing.ShutdownTracer)

	pool, err := connectPostgresWithRetry(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	closer.AddNamed("audit-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	repo := store.New(pool)
	if cfg.AutoMigrate {
		if err := repo.Migrate(ctx); err != nil {
			return nil, err
		}
	}

	mux := http.NewServeMux()
	handler := NewHandler(repo)
	mux.Handle("/audit/events", handler)
	mux.Handle("/metrics", pmetrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           ptracing.HTTPMiddleware("audit-http", pmetrics.HTTPMiddleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	closer.AddNamed("audit-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func connectPostgresWithRetry(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	const (
		attempts = 30
		delay    = time.Second
	)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err = pool.Ping(pingCtx)
			cancel()
			if err == nil {
				return pool, nil
			}
			pool.Close()
		}
		lastErr = err
		logger.Global().Warn("postgres is not ready for audit", "attempt", attempt, "max_attempts", attempts, "error", err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, lastErr
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting audit http server", "addr", a.server.Addr)
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
