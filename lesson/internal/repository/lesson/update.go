package lesson

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/lesson/internal/model"
	def "github.com/ossirdian/magistral/lesson/internal/repository"
)

func (r *repository) Update(ctx context.Context, schoolID, lessonID string, patch model.LessonPatch) (model.Lesson, error) {
	current, err := r.GetByID(ctx, schoolID, lessonID)
	if err != nil {
		return model.Lesson{}, err
	}

	candidate := current
	if patch.BranchIDSet {
		if patch.BranchID == nil {
			candidate.BranchID = ""
		} else {
			candidate.BranchID = *patch.BranchID
		}
	}
	if patch.StudentID != nil {
		candidate.StudentID = *patch.StudentID
	}
	if patch.InstructorID != nil {
		candidate.InstructorID = *patch.InstructorID
	}
	if patch.CarID != nil {
		candidate.CarID = *patch.CarID
	}
	if patch.StartAt != nil {
		candidate.StartAt = *patch.StartAt
	}
	if patch.EndAt != nil {
		candidate.EndAt = *patch.EndAt
	}

	conflict, err := r.hasConflict(ctx, schoolID, lessonConflictCandidate{
		StudentID:    candidate.StudentID,
		InstructorID: candidate.InstructorID,
		CarID:        candidate.CarID,
		StartAt:      candidate.StartAt,
		EndAt:        candidate.EndAt,
	}, &lessonID)
	if err != nil {
		return model.Lesson{}, err
	}
	if conflict {
		return model.Lesson{}, def.ErrConflict
	}

	builder := def.Builder().
		Update("lessons").
		Set("updated_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", lessonID).
		Where("deleted_at IS NULL")

	if patch.BranchIDSet {
		builder = builder.Set("branch_id", patch.BranchID)
	}
	if patch.StudentID != nil {
		builder = builder.Set("student_id", *patch.StudentID)
	}
	if patch.InstructorID != nil {
		builder = builder.Set("instructor_id", *patch.InstructorID)
	}
	if patch.CarID != nil {
		builder = builder.Set("car_id", *patch.CarID)
	}
	if patch.StartAt != nil {
		builder = builder.Set("start_at", *patch.StartAt)
	}
	if patch.EndAt != nil {
		builder = builder.Set("end_at", *patch.EndAt)
	}
	if patch.Status != nil {
		builder = builder.Set("status", *patch.Status)
	}
	if patch.CommentSet {
		builder = builder.Set("comment", patch.Comment)
	}

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.Lesson{}, err
	}

	var updatedID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return model.Lesson{}, def.ErrNotFound
		}
		return model.Lesson{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, updatedID)
}
