package car

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
)

func (r *repository) Update(ctx context.Context, schoolID, carID string, patch model.CarPatch) (model.Car, error) {
	repoPatch := converter.CarPatchFromModel(patch)
	builder := def.Builder().
		Update("cars").
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", carID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	changed := false

	if repoPatch.BranchID != nil {
		builder = builder.Set("branch_id", *repoPatch.BranchID)
		changed = true
	}
	if repoPatch.InstructorID.Set {
		builder = builder.Set("instructor_id", repoPatch.InstructorID.Value)
		changed = true
	}
	if repoPatch.PlateNumber != nil {
		builder = builder.Set("plate_number", *repoPatch.PlateNumber)
		changed = true
	}
	if repoPatch.Brand.Set {
		builder = builder.Set("brand", repoPatch.Brand.Value)
		changed = true
	}
	if repoPatch.Model.Set {
		builder = builder.Set("model", repoPatch.Model.Value)
		changed = true
	}
	if repoPatch.Color.Set {
		builder = builder.Set("color", repoPatch.Color.Value)
		changed = true
	}
	if repoPatch.Transmission != nil {
		builder = builder.Set("transmission", *repoPatch.Transmission)
		changed = true
	}
	if repoPatch.Status != nil {
		builder = builder.Set("status", *repoPatch.Status)
		changed = true
	}

	if !changed {
		return r.GetByID(ctx, schoolID, carID)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Car{}, err
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return model.Car{}, def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.Car{}, def.ErrNotFound
	}

	return r.GetByID(ctx, schoolID, carID)
}
