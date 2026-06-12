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

	query, args, err := deleteGeneratedInWindowSQL(schoolID, studentIDs, dateFrom, dateTo)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return def.NormalizeError(err)
}

func (r *repository) DeleteGeneratedInWindowByBranches(
	ctx context.Context,
	schoolID string,
	branchIDs []string,
	includeNullBranch bool,
	dateFrom,
	dateTo time.Time,
) error {
	query, args, err := deleteGeneratedInWindowByBranchesSQL(schoolID, branchIDs, includeNullBranch, dateFrom, dateTo)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return def.NormalizeError(err)
}

func deleteGeneratedInWindowSQL(schoolID string, studentIDs []string, dateFrom, dateTo time.Time) (string, []any, error) {
	return def.Builder().
		Update("lessons").
		Set("updated_at", sq.Expr("now()")).
		Set("deleted_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		Where("is_generated = ?", true).
		Where(sq.Eq{"student_id": studentIDs}).
		Where("start_at < ?", dateTo).
		Where("end_at > ?", dateFrom).
		ToSql()
}

func deleteGeneratedInWindowByBranchesSQL(
	schoolID string,
	branchIDs []string,
	includeNullBranch bool,
	dateFrom,
	dateTo time.Time,
) (string, []any, error) {
	builder := def.Builder().
		Update("lessons").
		Set("updated_at", sq.Expr("now()")).
		Set("deleted_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		Where("is_generated = ?", true).
		Where("start_at < ?", dateTo).
		Where("end_at > ?", dateFrom)

	if len(branchIDs) > 0 && includeNullBranch {
		builder = builder.Where(sq.Or{sq.Eq{"branch_id": branchIDs}, sq.Expr("branch_id IS NULL")})
	} else if len(branchIDs) > 0 {
		builder = builder.Where(sq.Eq{"branch_id": branchIDs})
	} else if includeNullBranch {
		builder = builder.Where("branch_id IS NULL")
	}

	return builder.ToSql()
}
