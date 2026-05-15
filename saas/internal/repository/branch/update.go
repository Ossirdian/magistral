package branch

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
)

func (r *repository) Update(ctx context.Context, branchID string, patch model.BranchPatch) (model.Branch, error) {
	repoPatch := converter.BranchPatchFromModel(patch)
	builder := def.Builder().
		Update("branches").
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", branchID).
		Where("deleted_at IS NULL")
	changed := false

	if repoPatch.SchoolID != nil {
		builder = builder.Set("school_id", *repoPatch.SchoolID)
		changed = true
	}
	if repoPatch.Name != nil {
		builder = builder.Set("name", *repoPatch.Name)
		changed = true
	}
	if repoPatch.Address.Set {
		builder = builder.Set("address", repoPatch.Address.Value)
		changed = true
	}
	if repoPatch.Status != nil {
		builder = builder.Set("status", *repoPatch.Status)
		changed = true
	}

	if !changed {
		return r.GetByID(ctx, branchID)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		r.logError(ctx, "update.build", err, "branch_id", branchID)
		return model.Branch{}, err
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "update.exec", err, "branch_id", branchID)
		return model.Branch{}, err
	}
	if tag.RowsAffected() == 0 {
		return model.Branch{}, def.ErrNotFound
	}

	return r.GetByID(ctx, branchID)
}
