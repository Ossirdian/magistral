//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ossirdian/magistral/platform/pkg/testcontainers/path"
	tcpostgres "github.com/ossirdian/magistral/platform/pkg/testcontainers/postgres"
	api "github.com/ossirdian/magistral/resources/internal/api/resources/v1"
	"github.com/ossirdian/magistral/resources/internal/migrator"
	"github.com/ossirdian/magistral/resources/internal/repository/car"
	"github.com/ossirdian/magistral/resources/internal/repository/instructor"
	carservice "github.com/ossirdian/magistral/resources/internal/service/car"
	instructorservice "github.com/ossirdian/magistral/resources/internal/service/instructor"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
)

func TestResourcesInstructorSmoke(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("testcontainers postgres smoke is executed in linux CI")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pg, err := tcpostgres.NewContainer(ctx, "resources", "resources_user", "resources_password")
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

	if err := migrator.NewMigrator(db, filepath.Join(repoRoot, "resources", "migrations")).Up(); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, pg.DSN())
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	defer pool.Close()

	schoolID := uuid.NewString()
	branchID := uuid.New()

	handler, err := api.NewServer(
		carservice.NewService(car.NewRepository(pool)),
		instructorservice.NewService(instructor.NewRepository(pool)),
		schoolID,
	)
	if err != nil {
		t.Fatalf("build api server: %v", err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := resourcesv1.NewClient(server.URL)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	createRes, err := client.CreateInstructor(ctx, &resourcesv1.InstructorCreateRequest{
		BranchID: branchID,
		FullName: "Smoke Test Instructor",
	})
	if err != nil {
		t.Fatalf("create instructor: %v", err)
	}

	createBody, ok := createRes.(*resourcesv1.InstructorResponse)
	if !ok {
		t.Fatalf("unexpected create response type %T", createRes)
	}

	if got := createBody.Data.FullName; got != "Smoke Test Instructor" {
		t.Fatalf("unexpected instructor name: %s", got)
	}

	listRes, err := client.ListInstructors(ctx, resourcesv1.ListInstructorsParams{})
	if err != nil {
		t.Fatalf("list instructors: %v", err)
	}

	listBody, ok := listRes.(*resourcesv1.InstructorListResponse)
	if !ok {
		t.Fatalf("unexpected list response type %T", listRes)
	}

	if len(listBody.Data) != 1 {
		t.Fatalf("unexpected instructors count: %d", len(listBody.Data))
	}

	if listBody.Data[0].ID != createBody.Data.ID {
		t.Fatalf("unexpected instructor id in list: %s", listBody.Data[0].ID)
	}
}
