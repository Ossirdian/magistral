package group

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) Update(ctx context.Context, schoolID, groupID string, patch model.GroupPatch) (model.Group, error) {
	builder := def.Builder().
		Update("groups").
		Set("updated_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", groupID).
		Where("deleted_at IS NULL")

	if patch.BranchID != nil {
		builder = builder.Set("branch_id", *patch.BranchID)
	}
	if patch.Name != nil {
		builder = builder.Set("name", *patch.Name)
	}

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.Group{}, err
	}

	var updatedID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return model.Group{}, def.ErrNotFound
		}
		return model.Group{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, updatedID)
}
