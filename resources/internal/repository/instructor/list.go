package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func (r *repository) List(ctx context.Context, schoolID string, filter model.InstructorFilter) ([]model.Instructor, model.PageMeta, error) {
	repoFilter := converter.InstructorFilterFromModel(filter)
	countBuilder := def.Builder().
		Select("count(*)").
		From("instructors i").
		Where("i.school_id = ?", schoolID).
		Where("i.deleted_at IS NULL")
	selectBuilder := def.Builder().
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
		Where("i.school_id = ?", schoolID).
		Where("i.deleted_at IS NULL")

	if repoFilter.BranchID != nil {
		countBuilder = countBuilder.Where("i.branch_id = ?", *repoFilter.BranchID)
		selectBuilder = selectBuilder.Where("i.branch_id = ?", *repoFilter.BranchID)
	}
	if repoFilter.Status != nil {
		countBuilder = countBuilder.Where("i.status = ?", *repoFilter.Status)
		selectBuilder = selectBuilder.Where("i.status = ?", *repoFilter.Status)
	}
	if repoFilter.AvailabilityStatus != nil {
		countBuilder = countBuilder.Where("i.availability_status = ?", *repoFilter.AvailabilityStatus)
		selectBuilder = selectBuilder.Where("i.availability_status = ?", *repoFilter.AvailabilityStatus)
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
		OrderBy("i.created_at DESC", "i.id DESC").
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

	items := make([]repomodel.Instructor, 0, limit)
	for rows.Next() {
		item, scanErr := scanInstructor(rows)
		if scanErr != nil {
			return nil, model.PageMeta{}, def.NormalizeError(scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, model.PageMeta{}, def.NormalizeError(err)
	}

	return converter.InstructorsToModel(items), converter.PageMetaToModel(repomodel.PageMeta{
		Limit:  limit,
		Offset: offset,
		Count:  total,
	}), nil
}
