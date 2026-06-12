package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
	auditpkg "github.com/ossirdian/magistral/platform/pkg/audit"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
)

type App struct {
	server                 *http.Server
	planningResultConsumer *planning.ResultConsumer
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

	lessonSvc := lessonservice.NewService(lessonrepo.NewRepository(pool))
	planningStatusStore := planning.NewStatusStore()

	planningResultConsumer, err := planning.NewResultConsumer(
		cfg.Planning().KafkaBootstrapServers(),
		cfg.Planning().KafkaResultTopic(),
		cfg.Planning().KafkaResultGroupID(),
		lessonSvc,
		planningStatusStore,
	)
	if err != nil {
		return nil, err
	}
	closer.AddNamed("lesson-planning-result-consumer", func(context.Context) error {
		return planningResultConsumer.Close()
	})

	handler, err := api.NewServer(lessonSvc, planningProducer, planningStatusStore)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/planning/requests/", planningStatusHandler(planningStatusStore))
	mux.Handle("/", handler)
	mux.Handle("/metrics", pmetrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           ptracing.HTTPMiddleware("lesson-http", pmetrics.HTTPMiddleware(auditpkg.Middleware("lesson", httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins())))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("lesson-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server, planningResultConsumer: planningResultConsumer}, nil
}

func planningStatusHandler(store *planning.StatusStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		schoolID, ok := httpauth.SchoolIDFromContext(r.Context())
		if !ok || schoolID == "" {
			writePlanningStatusJSON(w, http.StatusUnauthorized, map[string]any{
				"error": map[string]string{"code": "unauthorized", "message": "missing auth context"},
			})
			return
		}
		requestID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/planning/requests/"), "/")
		if requestID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		status, ok := store.Get(schoolID, requestID)
		if !ok {
			writePlanningStatusJSON(w, http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "not_found", "message": "planning request not found"},
			})
			return
		}
		writePlanningStatusJSON(w, http.StatusOK, map[string]any{"data": status})
	}
}

func writePlanningStatusJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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
	go func() {
		if err := a.planningResultConsumer.Run(ctx); err != nil {
			errCh <- err
		}
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
