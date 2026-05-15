package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	api "github.com/ossirdian/magistral/reporting/internal/api/reports/v1"
	"github.com/ossirdian/magistral/reporting/internal/config/env"
	"github.com/ossirdian/magistral/reporting/internal/httpauth"
	lessonrepo "github.com/ossirdian/magistral/reporting/internal/repository/lesson"
	paymentrepo "github.com/ossirdian/magistral/reporting/internal/repository/payment"
	resourcesrepo "github.com/ossirdian/magistral/reporting/internal/repository/resources"
	studentrepo "github.com/ossirdian/magistral/reporting/internal/repository/student"
	reportingservice "github.com/ossirdian/magistral/reporting/internal/service/reporting"
)

type App struct {
	server *http.Server
}

func New(ctx context.Context, cfg *env.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("reporting-tracer", ptracing.ShutdownTracer)

	studentPool, err := pgxpool.New(ctx, cfg.StudentPostgres().DSN())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("reporting-student-postgres-pool", func(context.Context) error {
		studentPool.Close()
		return nil
	})

	lessonPool, err := pgxpool.New(ctx, cfg.LessonPostgres().DSN())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("reporting-lesson-postgres-pool", func(context.Context) error {
		lessonPool.Close()
		return nil
	})

	resourcesPool, err := pgxpool.New(ctx, cfg.ResourcesPostgres().DSN())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("reporting-resources-postgres-pool", func(context.Context) error {
		resourcesPool.Close()
		return nil
	})

	paymentPool, err := pgxpool.New(ctx, cfg.PaymentPostgres().DSN())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("reporting-payment-postgres-pool", func(context.Context) error {
		paymentPool.Close()
		return nil
	})

	handler, err := api.NewServer(
		reportingservice.NewService(
			studentrepo.NewRepository(studentPool),
			lessonrepo.NewRepository(lessonPool),
			resourcesrepo.NewRepository(resourcesPool),
			paymentrepo.NewRepository(paymentPool),
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
		Handler:           ptracing.HTTPMiddleware("reporting-http", pmetrics.HTTPMiddleware(httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins()))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("reporting-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting reporting http server", "addr", a.server.Addr)
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
