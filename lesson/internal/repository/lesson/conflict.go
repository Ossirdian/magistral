package lesson

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	def "github.com/ossirdian/magistral/lesson/internal/repository"
)

func (r *repository) hasConflict(ctx context.Context, schoolID string, candidate lessonConflictCandidate, excludeID *string) (bool, error) {
	builder := def.Builder().
		Select("1").
		From("lessons").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		Where("status <> ?", "cancelled").
		Where("start_at < ?", candidate.EndAt).
		Where("end_at > ?", candidate.StartAt).
		Limit(1)

	if excludeID != nil {
		builder = builder.Where("id <> ?", *excludeID)
	}

	resourceOr := sq.Or{
		sq.Eq{"student_id": candidate.StudentID},
		sq.Eq{"instructor_id": candidate.InstructorID},
		sq.Eq{"car_id": candidate.CarID},
	}
	builder = builder.Where(resourceOr)

	query, args, err := builder.ToSql()
	if err != nil {
		return false, err
	}

	var marker int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&marker); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, def.NormalizeError(err)
	}

	return true, nil
}

type lessonConflictCandidate struct {
	StudentID    string
	InstructorID string
	CarID        string
	StartAt      time.Time
	EndAt        time.Time
}
