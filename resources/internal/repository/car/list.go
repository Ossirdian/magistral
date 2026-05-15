package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func (r *repository) List(ctx context.Context, schoolID string, filter model.CarFilter) ([]model.Car, model.PageMeta, error) {
	repoFilter := converter.CarFilterFromModel(filter)
	countBuilder := def.Builder().
		Select("count(*)").
		From("cars").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	selectBuilder := def.Builder().
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
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")

	if repoFilter.BranchID != nil {
		countBuilder = countBuilder.Where("branch_id = ?", *repoFilter.BranchID)
		selectBuilder = selectBuilder.Where("branch_id = ?", *repoFilter.BranchID)
	}
	if repoFilter.InstructorID != nil {
		countBuilder = countBuilder.Where("instructor_id = ?", *repoFilter.InstructorID)
		selectBuilder = selectBuilder.Where("instructor_id = ?", *repoFilter.InstructorID)
	}
	if repoFilter.Status != nil {
		countBuilder = countBuilder.Where("status = ?", *repoFilter.Status)
		selectBuilder = selectBuilder.Where("status = ?", *repoFilter.Status)
	}
	if repoFilter.AvailabilityStatus != nil {
		countBuilder = countBuilder.Where("availability_status = ?", *repoFilter.AvailabilityStatus)
		selectBuilder = selectBuilder.Where("availability_status = ?", *repoFilter.AvailabilityStatus)
	}
	limit := repoFilter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := repoFilter.Offset
	if offset < 0 {
		offset = 0
	}

	var total int
	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, model.PageMeta{}, err
	}
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, model.PageMeta{}, def.NormalizeError(err)
	}

	query, args, err := selectBuilder.
		OrderBy("created_at DESC", "id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, model.PageMeta{}, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, model.PageMeta{}, def.NormalizeError(err)
	}
	defer rows.Close()

	items := make([]repomodel.Car, 0, limit)
	for rows.Next() {
		item, scanErr := scanCar(rows)
		if scanErr != nil {
			return nil, model.PageMeta{}, def.NormalizeError(scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, model.PageMeta{}, def.NormalizeError(err)
	}

	return converter.CarsToModel(items), converter.PageMetaToModel(repomodel.PageMeta{
		Limit:  limit,
		Offset: offset,
		Count:  total,
	}), nil
}
