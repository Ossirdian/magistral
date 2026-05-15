package student

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) GetByID(ctx context.Context, schoolID, studentID string) (model.Student, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"branch_id::text",
			"group_id::text",
			"program_id::text",
			"full_name",
			"phone",
			"status",
			"preferred_weekdays",
			"preferred_time_windows",
			"manual_hours_completed",
			"created_at",
			"updated_at",
		).
		From("students").
		Where("school_id = ?", schoolID).
		Where("id = ?", studentID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Student{}, err
	}

	item, err := scanStudent(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, def.ErrNotFound
	}
	if err != nil {
		return model.Student{}, def.NormalizeError(err)
	}

	return item, nil
}
