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
	platformgrpc "github.com/ossirdian/magistral/platform/pkg/grpc/health"
	"github.com/ossirdian/magistral/platform/pkg/httpx"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	pmetrics "github.com/ossirdian/magistral/platform/pkg/prometheus"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	studentsv1grpc "github.com/ossirdian/magistral/shared/pkg/student/v1"
	grpcapi "github.com/ossirdian/magistral/student/internal/api/grpc/v1"
	api "github.com/ossirdian/magistral/student/internal/api/students/v1"
	"github.com/ossirdian/magistral/student/internal/config/env"
	"github.com/ossirdian/magistral/student/internal/httpauth"
	"github.com/ossirdian/magistral/student/internal/migrator"
	grouprepo "github.com/ossirdian/magistral/student/internal/repository/group"
	lessonrepo "github.com/ossirdian/magistral/student/internal/repository/lesson"
	programrepo "github.com/ossirdian/magistral/student/internal/repository/program"
	studentrepo "github.com/ossirdian/magistral/student/internal/repository/student"
	groupservice "github.com/ossirdian/magistral/student/internal/service/group"
	programservice "github.com/ossirdian/magistral/student/internal/service/program"
	studentservice "github.com/ossirdian/magistral/student/internal/service/student"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	httpServer *http.Server
	grpcServer *grpc.Server
	lis        net.Listener
}

func New(ctx context.Context, cfg *env.Config) (*App, error) {
	if err := ptracing.InitTracer(ctx, cfg.Tracing()); err != nil {
		return nil, err
	}
	closer.AddNamed("student-tracer", ptracing.ShutdownTracer)

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
	closer.AddNamed("student-postgres-pool", func(context.Context) error {
		pool.Close()
		return nil
	})

	lessonPool, err := pgxpool.New(ctx, cfg.LessonPostgres().DSN())
	if err != nil {
		return nil, err
	}
	closer.AddNamed("student-lesson-postgres-pool", func(context.Context) error {
		lessonPool.Close()
		return nil
	})

	programRepo := programrepo.NewRepository(pool)
	programSvc := programservice.NewService(programRepo)
	studentSvc := studentservice.NewService(studentrepo.NewRepository(pool), programRepo, lessonrepo.NewRepository(lessonPool))
	groupSvc := groupservice.NewService(grouprepo.NewRepository(pool))

	handler, err := api.NewServer(studentSvc, groupSvc, programSvc)
	if err != nil {
		return nil, err
	}

	grpcHandler, err := grpcapi.New(studentSvc)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen("tcp", cfg.GRPC().Addr())
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ptracing.UnaryServerInterceptor("student"),
			pmetrics.UnaryServerInterceptor(),
		),
	)
	studentsv1grpc.RegisterStudentServiceServer(grpcServer, grpcHandler)
	platformgrpc.RegisterService(grpcServer)
	reflection.Register(grpcServer)

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/metrics", pmetrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	httpServer := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           ptracing.HTTPMiddleware("student-http", pmetrics.HTTPMiddleware(httpx.WithCORS(httpauth.Middleware(mux), cfg.HTTP().AllowedOrigins()))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("student-grpc-server", func(closeCtx context.Context) error {
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

	closer.AddNamed("student-http-server", func(closeCtx context.Context) error {
		return httpServer.Shutdown(closeCtx)
	})

	return &App{httpServer: httpServer, grpcServer: grpcServer, lis: lis}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)
	go func() {
		logger.Global().Info("starting student grpc server", "addr", a.lis.Addr().String())
		errCh <- a.grpcServer.Serve(a.lis)
	}()
	go func() {
		logger.Global().Info("starting student http server", "addr", a.httpServer.Addr)
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
