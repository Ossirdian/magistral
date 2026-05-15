package instructor

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
)

func (r *repository) GetByID(ctx context.Context, schoolID, instructorID string) (model.Instructor, error) {
	query, args, err := def.Builder().
		Select(
			"i.id::text",
			"i.school_id::text",
			"i.branch_id::text",
			"i.full_name",
			"i.phone",
			"i.email",
			"i.status",
			"i.availability_status",
			`(
				SELECT c.id::text
				FROM cars c
				WHERE c.instructor_id = i.id
				  AND c.school_id = i.school_id
				  AND c.deleted_at IS NULL
				ORDER BY c.updated_at DESC, c.created_at DESC
				LIMIT 1
			)`,
			"i.created_at",
			"i.updated_at",
		).
		From("instructors i").
		Where("i.id = ?", instructorID).
		Where("i.school_id = ?", schoolID).
		Where("i.deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Instructor{}, err
	}

	item, err := scanInstructor(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Instructor{}, def.ErrNotFound
		}
		return model.Instructor{}, def.NormalizeError(err)
	}

	return converter.InstructorToModel(item), nil
}
