package branch

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
)

func (r *repository) GetByID(ctx context.Context, branchID string) (model.Branch, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"name",
			"address",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		From("branches").
		Where("id = ?", branchID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		r.logError(ctx, "get_by_id.build", err, "branch_id", branchID)
		return model.Branch{}, err
	}

	item, err := scanBranch(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Branch{}, def.ErrNotFound
		}

		err = def.NormalizeError(err)
		r.logError(ctx, "get_by_id.query", err, "branch_id", branchID)
		return model.Branch{}, err
	}

	return converter.BranchToModel(item), nil
}
