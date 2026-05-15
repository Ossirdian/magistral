package student

import (
	"context"

	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) Create(ctx context.Context, schoolID string, input model.StudentCreate) (model.Student, error) {
	builder := def.Builder().
		Insert("students").
		Columns("school_id", "branch_id", "group_id", "program_id", "full_name", "phone", "preferred_weekdays", "preferred_time_windows", "manual_hours_completed")

	manualHours := 0.0
	if input.ManualHoursCompleted != nil {
		manualHours = *input.ManualHoursCompleted
	}
	values := []any{
		schoolID,
		input.BranchID,
		input.GroupID,
		input.ProgramID,
		input.FullName,
		input.Phone,
		preferredWeekdaysToStrings(input.PreferredWeekdays),
		preferredTimeWindowsToStrings(input.PreferredTimeWindows),
		manualHours,
	}
	if input.Status != nil {
		builder = builder.Columns("status")
		values = append(values, *input.Status)
	}

	query, args, err := builder.Values(values...).Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.Student{}, err
	}

	var studentID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&studentID); err != nil {
		return model.Student{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, studentID)
}
