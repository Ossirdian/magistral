package group

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) GetByID(ctx context.Context, schoolID, groupID string) (model.Group, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"branch_id::text",
			"name",
			"created_at",
			"updated_at",
		).
		From("groups").
		Where("school_id = ?", schoolID).
		Where("id = ?", groupID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Group{}, err
	}

	item, err := scanGroup(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Group{}, def.ErrNotFound
	}
	if err != nil {
		return model.Group{}, def.NormalizeError(err)
	}

	return item, nil
}
