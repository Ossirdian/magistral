package branch

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/repository"
	"github.com/ossirdian/magistral/saas/internal/repository/converter"
	repomodel "github.com/ossirdian/magistral/saas/internal/repository/model"
)

func (r *repository) List(ctx context.Context, filter model.BranchFilter) ([]model.Branch, model.PageMeta, error) {
	repoFilter := converter.BranchFilterFromModel(filter)
	countBuilder := def.Builder().
		Select("count(*)").
		From("branches").
		Where("deleted_at IS NULL")
	selectBuilder := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"name",
			"address",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		From("branches").
		Where("deleted_at IS NULL")

	if repoFilter.SchoolID != nil {
		countBuilder = countBuilder.Where("school_id = ?", *repoFilter.SchoolID)
		selectBuilder = selectBuilder.Where("school_id = ?", *repoFilter.SchoolID)
	}
	if repoFilter.Status != nil {
		countBuilder = countBuilder.Where("status = ?", *repoFilter.Status)
		selectBuilder = selectBuilder.Where("status = ?", *repoFilter.Status)
	}

	limit := repoFilter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := repoFilter.Offset
	if offset < 0 {
		offset = 0
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		r.logError(ctx, "list.count.build", err)
		return nil, model.PageMeta{}, err
	}

	var total int
	if err = r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "list.count.query", err)
		return nil, model.PageMeta{}, err
	}

	query, args, err := selectBuilder.
		OrderBy("created_at DESC", "id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		r.logError(ctx, "list.select.build", err)
		return nil, model.PageMeta{}, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "list.select.query", err)
		return nil, model.PageMeta{}, err
	}
	defer rows.Close()

	items := make([]repomodel.Branch, 0, limit)
	for rows.Next() {
		item, scanErr := scanBranch(rows)
		if scanErr != nil {
			scanErr = def.NormalizeError(scanErr)
			r.logError(ctx, "list.scan", scanErr)
			return nil, model.PageMeta{}, scanErr
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		err = def.NormalizeError(err)
		r.logError(ctx, "list.rows", err)
		return nil, model.PageMeta{}, err
	}

	return converter.BranchesToModel(items), converter.PageMetaToModel(repomodel.PageMeta{
		Limit:  limit,
		Offset: offset,
		Count:  total,
	}), nil
}
