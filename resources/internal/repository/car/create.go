package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, schoolID string, input model.CarCreate) (model.Car, error) {
	repoInput := converter.CarCreateFromModel(input)
	columns := []string{"school_id", "branch_id", "instructor_id", "plate_number", "brand", "model", "color"}
	values := []any{schoolID, repoInput.BranchID, repoInput.InstructorID, repoInput.PlateNumber, repoInput.Brand, repoInput.Model, repoInput.Color}

	if repoInput.Transmission != nil {
		columns = append(columns, "transmission")
		values = append(values, *repoInput.Transmission)
	}
	if repoInput.Status != nil {
		columns = append(columns, "status")
		values = append(values, *repoInput.Status)
	}

	builder := def.Builder().
		Insert("cars").
		Columns(columns...).
		Values(values...)

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.Car{}, err
	}

	var carID string
	if err = r.pool.QueryRow(ctx, query, args...).Scan(&carID); err != nil {
		return model.Car{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, carID)
}
