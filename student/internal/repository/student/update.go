package student

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) Update(ctx context.Context, schoolID, studentID string, patch model.StudentPatch) (model.Student, error) {
	builder := def.Builder().
		Update("students").
		Set("updated_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", studentID).
		Where("deleted_at IS NULL")

	if patch.BranchID != nil {
		builder = builder.Set("branch_id", *patch.BranchID)
	}
	if patch.GroupIDSet {
		builder = builder.Set("group_id", patch.GroupID)
	}
	if patch.ProgramID != nil {
		builder = builder.Set("program_id", *patch.ProgramID)
	}
	if patch.FullName != nil {
		builder = builder.Set("full_name", *patch.FullName)
	}
	if patch.Phone != nil {
		builder = builder.Set("phone", *patch.Phone)
	}
	if patch.Status != nil {
		builder = builder.Set("status", *patch.Status)
	}
	builder = builder.Set("preferred_weekdays", preferredWeekdaysToStrings(patch.PreferredWeekdays))
	builder = builder.Set("preferred_time_windows", preferredTimeWindowsToStrings(patch.PreferredTimeWindows))
	if patch.ManualHoursCompleted != nil {
		builder = builder.Set("manual_hours_completed", *patch.ManualHoursCompleted)
	}

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.Student{}, err
	}

	var updatedID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return model.Student{}, def.ErrNotFound
		}
		return model.Student{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, updatedID)
}
