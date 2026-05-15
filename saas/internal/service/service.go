package service

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

type SchoolService interface {
	Create(ctx context.Context, input model.SchoolCreate) (model.School, error)
	BootstrapWithAdmin(ctx context.Context, input model.SchoolCreate, admin model.SchoolAdminBootstrap) (model.School, string, error)
	Get(ctx context.Context, schoolID string) (model.School, error)
	List(ctx context.Context, filter model.SchoolFilter) ([]model.School, model.PageMeta, error)
	Update(ctx context.Context, schoolID string, patch model.SchoolPatch) (model.School, error)
	Delete(ctx context.Context, schoolID string) error
}

type BranchService interface {
	Create(ctx context.Context, input model.BranchCreate) (model.Branch, error)
	Get(ctx context.Context, branchID string) (model.Branch, error)
	List(ctx context.Context, filter model.BranchFilter) ([]model.Branch, model.PageMeta, error)
	Update(ctx context.Context, branchID string, patch model.BranchPatch) (model.Branch, error)
	Delete(ctx context.Context, branchID string) error
}
