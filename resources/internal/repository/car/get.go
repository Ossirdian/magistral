package car

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
)

func (r *repository) GetByID(ctx context.Context, schoolID, carID string) (model.Car, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"branch_id::text",
			"instructor_id::text",
			"plate_number",
			"brand",
			"model",
			"color",
			"transmission",
			"status",
			"availability_status",
			"created_at",
			"updated_at",
		).
		From("cars").
		Where("id = ?", carID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Car{}, err
	}

	item, err := scanCar(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Car{}, def.ErrNotFound
		}
		return model.Car{}, def.NormalizeError(err)
	}

	return converter.CarToModel(item), nil
}
