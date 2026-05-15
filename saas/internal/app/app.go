package app

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	adminapi "github.com/ossirdian/magistral/saas/internal/api/admin"
	api "github.com/ossirdian/magistral/saas/internal/api/saasadmin/v1"
	iamclient "github.com/ossirdian/magistral/saas/internal/client/iam"
	cfgenv "github.com/ossirdian/magistral/saas/internal/config/env"
	"github.com/ossirdian/magistral/saas/internal/httpauth"
	"github.com/ossirdian/magistral/saas/internal/migrator"
	"github.com/ossirdian/magistral/saas/internal/repository/branch"
	"github.com/ossirdian/magistral/saas/internal/repository/school"
	branchservice "github.com/ossirdian/magistral/saas/internal/service/branch"
	schoolservice "github.com/ossirdian/magistral/saas/internal/service/school"
)

type App struct {
	server *http.Server
}

func New(ctx context.Context, cfg *cfgenv.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("saas-tracer", ptracing.ShutdownTracer)

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
	closer.AddNamed("saas-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	var schoolIAMClient schoolservice.IAMClient
	if cfg.IAM().BaseURL() != "" {
		schoolIAMClient = iamclient.NewHTTPClient(
			cfg.IAM().BaseURL(),
			&http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					DialContext: (&net.Dialer{Timeout: 3 * time.Second}).DialContext,
				},
			},
		)
	}

	schoolSvc := schoolservice.NewService(school.NewRepository(pool, logger.Global()), schoolIAMClient, logger.Global())
	branchSvc := branchservice.NewService(branch.NewRepository(pool, logger.Global()))

	handler, err := api.NewServer(schoolSvc, branchSvc)
	if err != nil {
		return nil, err
	}

	adminHandler, err := adminapi.NewSchoolBootstrapHandler(schoolSvc)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/admin/bootstrap-school", adminHandler)
	mux.Handle("/", handler)
	mux.Handle("/metrics", pmetrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handlerWithMiddleware := ptracing.HTTPMiddleware(
		"saas-http",
		pmetrics.HTTPMiddleware(httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins())),
	)

	server := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           handlerWithMiddleware,
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("saas-http-server", func(closeCtx context.Context) error {
		return server.Shutdown(closeCtx)
	})

	return &App{server: server}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Global().Info("starting saas http server", "addr", a.server.Addr)
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
