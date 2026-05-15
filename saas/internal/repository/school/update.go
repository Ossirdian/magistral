package school

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
)

func (r *repository) Update(ctx context.Context, schoolID string, patch model.SchoolPatch) (model.School, error) {
	repoPatch := converter.SchoolPatchFromModel(patch)
	builder := def.Builder().
		Update("schools").
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", schoolID).
		Where("deleted_at IS NULL")
	changed := false

	if repoPatch.Name != nil {
		builder = builder.Set("name", *repoPatch.Name)
		changed = true
	}
	if repoPatch.Status != nil {
		builder = builder.Set("status", *repoPatch.Status)
		changed = true
	}

	if !changed {
		return r.GetByID(ctx, schoolID)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		r.logError(ctx, "update.build", err, "school_id", schoolID)
		return model.School{}, err
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "update.exec", err, "school_id", schoolID)
		return model.School{}, err
	}
	if tag.RowsAffected() == 0 {
		return model.School{}, def.ErrNotFound
	}

	return r.GetByID(ctx, schoolID)
}
