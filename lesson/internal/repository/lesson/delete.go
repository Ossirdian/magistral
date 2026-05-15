package lesson

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	def "github.com/ossirdian/magistral/lesson/internal/repository"
)

func (r *repository) Delete(ctx context.Context, schoolID, lessonID string) error {
	query, args, err := def.Builder().
		Update("lessons").
		Set("updated_at", sq.Expr("now()")).
		Set("deleted_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", lessonID).
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

func (r *repository) DeleteGeneratedInWindow(ctx context.Context, schoolID string, studentIDs []string, dateFrom, dateTo time.Time) error {
	if len(studentIDs) == 0 {
		return nil
	}

	query, args, err := def.Builder().
		Update("lessons").
		Set("updated_at", sq.Expr("now()")).
		Set("deleted_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		Where("is_generated = ?", true).
		Where("student_id IN (?)", studentIDs).
		Where("start_at < ?", dateTo).
		Where("end_at > ?", dateFrom).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return def.NormalizeError(err)
}
