package lesson

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/lesson/internal/model"
	def "github.com/ossirdian/magistral/lesson/internal/repository"
)

func (r *repository) GetByID(ctx context.Context, schoolID, lessonID string) (model.Lesson, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"COALESCE(branch_id::text, '')",
			"student_id::text",
			"instructor_id::text",
			"car_id::text",
			"start_at",
			"end_at",
			"status",
			"comment",
			"is_generated",
			"created_at",
			"updated_at",
		).
		From("lessons").
		Where("school_id = ?", schoolID).
		Where("id = ?", lessonID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Lesson{}, err
	}

	item, err := scanLesson(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Lesson{}, def.ErrNotFound
	}
	if err != nil {
		return model.Lesson{}, def.NormalizeError(err)
	}

	return item, nil
}
