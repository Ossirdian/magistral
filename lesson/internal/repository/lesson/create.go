package lesson

import (
	"context"

	"github.com/ossirdian/magistral/lesson/internal/model"
	def "github.com/ossirdian/magistral/lesson/internal/repository"
)

func (r *repository) Create(ctx context.Context, schoolID string, input model.LessonCreate) (model.Lesson, error) {
	conflict, err := r.hasConflict(ctx, schoolID, lessonConflictCandidate{
		StudentID:    input.StudentID,
		InstructorID: input.InstructorID,
		CarID:        input.CarID,
		StartAt:      input.StartAt,
		EndAt:        input.EndAt,
	}, nil)
	if err != nil {
		return model.Lesson{}, err
	}
	if conflict {
		return model.Lesson{}, def.ErrConflict
	}

	query, args, err := def.Builder().
		Insert("lessons").
		Columns("school_id", "branch_id", "student_id", "instructor_id", "car_id", "start_at", "end_at", "status", "comment", "is_generated").
		Values(schoolID, input.BranchID, input.StudentID, input.InstructorID, input.CarID, input.StartAt, input.EndAt, *input.Status, input.Comment, input.IsGenerated).
		Suffix("RETURNING id::text").
		ToSql()
	if err != nil {
		return model.Lesson{}, err
	}

	var lessonID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&lessonID); err != nil {
		return model.Lesson{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, lessonID)
}
