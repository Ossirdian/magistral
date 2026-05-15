package group

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

func (r *repository) List(ctx context.Context, schoolID string, filter model.GroupFilter) ([]model.Group, model.PageMeta, error) {
	countBuilder := def.Builder().
		Select("count(*)").
		From("groups").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	selectBuilder := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"branch_id::text",
			"name",
			"created_at",
			"updated_at",
		).
		From("groups").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")

	if filter.BranchID != nil {
		countBuilder = countBuilder.Where("branch_id = ?", *filter.BranchID)
		selectBuilder = selectBuilder.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Search != nil {
		pattern := "%" + strings.TrimSpace(*filter.Search) + "%"
		countBuilder = countBuilder.Where("name ILIKE ?", pattern)
		selectBuilder = selectBuilder.Where("name ILIKE ?", pattern)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, model.PageMeta{}, err
	}

	var total int
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

	items := make([]model.Group, 0, limit)
	for rows.Next() {
		item, scanErr := scanGroup(rows)
		if scanErr != nil {
			return nil, model.PageMeta{}, def.NormalizeError(scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, model.PageMeta{}, def.NormalizeError(err)
	}

	return items, model.PageMeta{Limit: limit, Offset: offset, Count: total}, nil
}
