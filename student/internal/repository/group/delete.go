package group

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) Delete(ctx context.Context, schoolID, groupID string) error {
	query, args, err := def.Builder().
		Update("groups").
		Set("updated_at", sq.Expr("now()")).
		Set("deleted_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", groupID).
		Where("deleted_at IS NULL").
		Suffix("RETURNING id::text").
		ToSql()
	if err != nil {
		return err
	}

	var deletedID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&deletedID); err != nil {
		if err == pgx.ErrNoRows {
			return def.ErrNotFound
		}
		return def.NormalizeError(err)
	}
	return nil
}
