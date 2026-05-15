package student

import (
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/student/internal/model"
)

func scanStudent(row pgx.Row) (model.Student, error) {
	var item model.Student
	var groupID *string
	var programID *string
	var preferredWeekdays []string
	var preferredTimeWindows []string
	err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.BranchID,
		&groupID,
		&programID,
		&item.FullName,
		&item.Phone,
		&item.Status,
		&preferredWeekdays,
		&preferredTimeWindows,
		&item.ManualHoursCompleted,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	item.GroupID = groupID
	item.ProgramID = programID
	item.PreferredWeekdays = make([]model.StudentPreferredWeekday, 0, len(preferredWeekdays))
	for _, value := range preferredWeekdays {
		item.PreferredWeekdays = append(item.PreferredWeekdays, model.StudentPreferredWeekday(value))
	}
	item.PreferredTimeWindows = make([]model.StudentPreferredTimeWindow, 0, len(preferredTimeWindows))
	for _, value := range preferredTimeWindows {
		item.PreferredTimeWindows = append(item.PreferredTimeWindows, model.StudentPreferredTimeWindow(value))
	}
	return item, err
}
