//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"net"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	apiv1 "github.com/ossirdian/magistral/iam/internal/api/grpc/v1"
	"github.com/ossirdian/magistral/iam/internal/config"
	"github.com/ossirdian/magistral/iam/internal/migrator"
	"github.com/ossirdian/magistral/iam/internal/repository"
	iamservice "github.com/ossirdian/magistral/iam/internal/service/iam"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	"github.com/ossirdian/magistral/platform/pkg/testcontainers/path"
	tcpostgres "github.com/ossirdian/magistral/platform/pkg/testcontainers/postgres"
	iamv1 "github.com/ossirdian/magistral/shared/pkg/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type authConfig struct{}

func (authConfig) JWTSecret() string              { return "integration-secret" }
func (authConfig) Issuer() string                 { return "integration-iam" }
func (authConfig) AccessTokenTTL() time.Duration  { return 15 * time.Minute }
func (authConfig) RefreshTokenTTL() time.Duration { return 24 * time.Hour }
func (authConfig) ResetTokenTTL() time.Duration   { return time.Hour }

var _ config.AuthConfig = authConfig{}

func TestIAMUserAuthSmoke(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("testcontainers postgres smoke is executed in linux CI")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pg, err := tcpostgres.NewContainer(ctx, "iam", "iam_user", "iam_password")
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	defer func() { _ = pg.Terminate(context.Background()) }()

	db, err := sql.Open("pgx", pg.DSN())
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer func() { _ = db.Close() }()

	repoRoot, err := path.RepoRoot()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	if err := migrator.NewMigrator(db, filepath.Join(repoRoot, "iam", "migrations")).Up(); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, pg.DSN())
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	defer pool.Close()

	handler, err := apiv1.New(iamservice.NewService(
		repository.NewPostgresRepository(pool, logger.Global()),
		authConfig{},
		logger.Global(),
	))
	if err != nil {
		t.Fatalf("build iam api: %v", err)
	}

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	iamv1.RegisterIAMServiceServer(server, handler)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	defer func() { _ = conn.Close() }()

	client := iamv1.NewIAMServiceClient(conn)
	schoolID := uuid.NewString()

	user, err := client.CreateUser(ctx, &iamv1.CreateUserRequest{
		SchoolId: &schoolID,
		Email:    "admin@example.com",
		Password: "Passw0rd!",
		Role:     iamv1.UserRole_USER_ROLE_ADMIN,
		Status:   iamv1.UserStatus_USER_STATUS_ACTIVE,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.GetId() == "" {
		t.Fatal("expected user id")
	}

	login, err := client.Login(ctx, &iamv1.LoginRequest{
		Email:    "admin@example.com",
		Password: "Passw0rd!",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.GetAccessToken() == "" || login.GetRefreshToken() == "" {
		t.Fatal("expected access and refresh tokens")
	}

	validation, err := client.ValidateToken(ctx, &iamv1.ValidateTokenRequest{AccessToken: login.GetAccessToken()})
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if !validation.GetIsValid() || validation.GetUserId() != user.GetId() {
		t.Fatalf("unexpected validation response: %+v", validation)
	}

	refreshed, err := client.RefreshSession(ctx, &iamv1.RefreshSessionRequest{RefreshToken: login.GetRefreshToken()})
	if err != nil {
		t.Fatalf("refresh session: %v", err)
	}
	if refreshed.GetAccessToken() == "" || refreshed.GetRefreshToken() == "" {
		t.Fatal("expected refreshed tokens")
	}
}
