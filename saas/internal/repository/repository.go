package repository

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

type SchoolRepository interface {
	GetByID(ctx context.Context, schoolID string) (model.School, error)
	List(ctx context.Context, filter model.SchoolFilter) ([]model.School, model.PageMeta, error)
	Create(ctx context.Context, input model.SchoolCreate) (model.School, error)
	Update(ctx context.Context, schoolID string, patch model.SchoolPatch) (model.School, error)
	Delete(ctx context.Context, schoolID string) error
}

type BranchRepository interface {
	GetByID(ctx context.Context, branchID string) (model.Branch, error)
	List(ctx context.Context, filter model.BranchFilter) ([]model.Branch, model.PageMeta, error)
	Create(ctx context.Context, input model.BranchCreate) (model.Branch, error)
	Update(ctx context.Context, branchID string, patch model.BranchPatch) (model.Branch, error)
	Delete(ctx context.Context, branchID string) error
}
