package app

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"time"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	envoyauthv1 "github.com/ossirdian/magistral/iam/internal/api/envoyauth/v1"
	apiv1 "github.com/ossirdian/magistral/iam/internal/api/grpc/v1"
	httpapi "github.com/ossirdian/magistral/iam/internal/api/http"
	"github.com/ossirdian/magistral/iam/internal/config/env"
	"github.com/ossirdian/magistral/iam/internal/migrator"
	"github.com/ossirdian/magistral/iam/internal/model"
	"github.com/ossirdian/magistral/iam/internal/repository"
	svcpkg "github.com/ossirdian/magistral/iam/internal/service"
	iamservice "github.com/ossirdian/magistral/iam/internal/service/iam"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	platformgrpc "github.com/ossirdian/magistral/platform/pkg/grpc/health"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	iamv1 "github.com/ossirdian/magistral/shared/pkg/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	lis        net.Listener
}

func New(ctx context.Context, cfg *env.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("iam-tracer", ptracing.ShutdownTracer)

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
	closer.AddNamed("iam-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	service := iamservice.NewService(
		repository.NewPostgresRepository(pool, logger.Global()),
		cfg.Auth(),
		logger.Global(),
	)
	if err := bootstrapSuperAdmin(ctx, service, cfg); err != nil {
		return nil, err
	}

	grpcHandler, err := apiv1.New(service)
	if err != nil {
		return nil, err
	}

	envoyAuthHandler, err := envoyauthv1.New(service)
	if err != nil {
		return nil, err
	}

	httpHandler, err := httpapi.New(service)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen("tcp", cfg.GRPC().Addr())
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ptracing.UnaryServerInterceptor("iam"),
			pmetrics.UnaryServerInterceptor(),
		),
	)
	iamv1.RegisterIAMServiceServer(grpcServer, grpcHandler)
	authv3.RegisterAuthorizationServer(grpcServer, envoyAuthHandler)
	platformgrpc.RegisterService(grpcServer)
	reflection.Register(grpcServer)

	httpMux := http.NewServeMux()
	httpMux.Handle("/", httpHandler.Handler())
	httpMux.Handle("/metrics", pmetrics.Handler())

	httpServer := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           ptracing.HTTPMiddleware("iam-http", pmetrics.HTTPMiddleware(httpx.WithCORS(httpMux, cfg.HTTP().AllowedOrigins()))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("iam-grpc-server", func(closeCtx context.Context) error {
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
			return nil
		case <-closeCtx.Done():
			grpcServer.Stop()
			return closeCtx.Err()
		}
	})

	closer.AddNamed("iam-http-server", func(closeCtx context.Context) error {
		return httpServer.Shutdown(closeCtx)
	})

	return &App{grpcServer: grpcServer, httpServer: httpServer, lis: lis}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)
	go func() {
		logger.Global().Info("starting iam grpc server", "addr", a.lis.Addr().String())
		errCh <- a.grpcServer.Serve(a.lis)
	}()
	go func() {
		logger.Global().Info("starting iam http server", "addr", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func bootstrapSuperAdmin(ctx context.Context, service *iamservice.Service, cfg *env.Config) error {
	email := cfg.Bootstrap().SuperAdminEmail()
	password := cfg.Bootstrap().SuperAdminPassword()
	if email == "" || password == "" {
		return nil
	}

	_, err := service.GetUserByEmail(ctx, email)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	_, err = service.CreateUser(ctx, model.UserCreate{
		Email:    email,
		Password: password,
		Role:     model.UserRoleSuperAdmin,
		Status:   model.UserStatusActive,
	})
	if err != nil && !errors.Is(err, repository.ErrConflict) && !errors.Is(err, svcpkg.ErrInvalidArgument) {
		return err
	}

	logger.Global().Info("bootstrap super admin is ready", "email", email)
	return nil
}
