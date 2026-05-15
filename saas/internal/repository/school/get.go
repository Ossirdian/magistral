package school

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
)

func (r *repository) GetByID(ctx context.Context, schoolID string) (model.School, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"name",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		From("schools").
		Where("id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		r.logError(ctx, "get_by_id.build", err, "school_id", schoolID)
		return model.School{}, err
	}

	item, err := scanSchool(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.School{}, def.ErrNotFound
		}

		err = def.NormalizeError(err)
		r.logError(ctx, "get_by_id.query", err, "school_id", schoolID)
		return model.School{}, err
	}

	return converter.SchoolToModel(item), nil
}
