package instructor

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	def "github.com/ossirdian/magistral/resources/internal/repository"
)

func (r *repository) Delete(ctx context.Context, schoolID, instructorID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return def.NormalizeError(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	deleteInstructorQuery, deleteInstructorArgs, err := def.Builder().
		Update("instructors").
		Set("deleted_at", sq.Expr("now()")).
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", instructorID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, deleteInstructorQuery, deleteInstructorArgs...)
	if err != nil {
		return def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return def.ErrNotFound
	}

	unassignCarsQuery, unassignCarsArgs, err := def.Builder().
		Update("cars").
		Set("instructor_id", nil).
		Set("updated_at", sq.Expr("now()")).
		Where("instructor_id = ?", instructorID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, unassignCarsQuery, unassignCarsArgs...); err != nil {
		return def.NormalizeError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return def.NormalizeError(err)
	}

	return nil
}
