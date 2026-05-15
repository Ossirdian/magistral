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

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	"github.com/ossirdian/magistral/platform/pkg/testcontainers/path"
	tcpostgres "github.com/ossirdian/magistral/platform/pkg/testcontainers/postgres"
	api "github.com/ossirdian/magistral/saas/internal/api/saasadmin/v1"
	"github.com/ossirdian/magistral/saas/internal/migrator"
	"github.com/ossirdian/magistral/saas/internal/repository/branch"
	"github.com/ossirdian/magistral/saas/internal/repository/school"
	branchservice "github.com/ossirdian/magistral/saas/internal/service/branch"
	schoolservice "github.com/ossirdian/magistral/saas/internal/service/school"
	saasadminv1 "github.com/ossirdian/magistral/shared/pkg/saasadmin/v1"
)

func TestSaaSSchoolBranchSmoke(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("testcontainers postgres smoke is executed in linux CI")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pg, err := tcpostgres.NewContainer(ctx, "saas", "saas_user", "saas_password")
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

	if err := migrator.NewMigrator(db, filepath.Join(repoRoot, "saas", "migrations")).Up(); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, pg.DSN())
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	defer pool.Close()

	handler, err := api.NewServer(
		schoolservice.NewService(school.NewRepository(pool, logger.Global()), nil, logger.Global()),
		branchservice.NewService(branch.NewRepository(pool, logger.Global())),
	)
	if err != nil {
		t.Fatalf("build api server: %v", err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := saasadminv1.NewClient(server.URL)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	createSchoolRes, err := client.CreateSchool(ctx, &saasadminv1.SchoolCreateRequest{
		Name: "Smoke School",
	})
	if err != nil {
		t.Fatalf("create school: %v", err)
	}

	createSchoolBody, ok := createSchoolRes.(*saasadminv1.SchoolResponse)
	if !ok {
		t.Fatalf("unexpected create school response type %T", createSchoolRes)
	}

	createBranchRes, err := client.CreateBranch(ctx, &saasadminv1.BranchCreateRequest{
		SchoolID: createSchoolBody.Data.ID,
		Name:     "Main Branch",
	})
	if err != nil {
		t.Fatalf("create branch: %v", err)
	}

	createBranchBody, ok := createBranchRes.(*saasadminv1.BranchResponse)
	if !ok {
		t.Fatalf("unexpected create branch response type %T", createBranchRes)
	}

	listRes, err := client.ListBranches(ctx, saasadminv1.ListBranchesParams{
		SchoolID: saasadminv1.NewOptUUID(createSchoolBody.Data.ID),
	})
	if err != nil {
		t.Fatalf("list branches: %v", err)
	}

	listBody, ok := listRes.(*saasadminv1.BranchListResponse)
	if !ok {
		t.Fatalf("unexpected list branches response type %T", listRes)
	}

	if len(listBody.Data) != 1 {
		t.Fatalf("unexpected branches count: %d", len(listBody.Data))
	}

	if listBody.Data[0].ID != createBranchBody.Data.ID {
		t.Fatalf("unexpected branch id in list: %s", listBody.Data[0].ID)
	}
}
