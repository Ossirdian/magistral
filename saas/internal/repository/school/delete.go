package school

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	def "github.com/ossirdian/magistral/saas/internal/repository"
)

func (r *repository) Delete(ctx context.Context, schoolID string) error {
	query, args, err := def.Builder().
		Update("schools").
		Set("deleted_at", sq.Expr("now()")).
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		r.logError(ctx, "delete.build", err, "school_id", schoolID)
		return err
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "delete.exec", err, "school_id", schoolID)
		return err
	}
	if tag.RowsAffected() == 0 {
		return def.ErrNotFound
	}

	return nil
}
