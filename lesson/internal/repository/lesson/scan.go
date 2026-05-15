package lesson

import (
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/lesson/internal/model"
)

func scanLesson(row pgx.Row) (model.Lesson, error) {
	var item model.Lesson
	err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.BranchID,
		&item.StudentID,
		&item.InstructorID,
		&item.CarID,
		&item.StartAt,
		&item.EndAt,
		&item.Status,
		&item.Comment,
		&item.IsGenerated,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}
