package car

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	def "github.com/ossirdian/magistral/resources/internal/repository"
)

func (r *repository) Delete(ctx context.Context, schoolID, carID string) error {
	query, args, err := def.Builder().
		Update("cars").
		Set("deleted_at", sq.Expr("now()")).
		Set("updated_at", sq.Expr("now()")).
		Set("instructor_id", nil).
		Where("id = ?", carID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return def.ErrNotFound
	}

	return nil
}
