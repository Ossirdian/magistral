package program

import (
	"context"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/student/internal/model"
	def "github.com/ossirdian/magistral/student/internal/repository"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}

func (r *repository) GetByID(ctx context.Context, schoolID, programID string) (model.StudyProgram, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"name",
			"transmission",
			"total_hours",
			"created_at",
			"updated_at",
		).
		From("study_programs").
		Where("school_id = ?", schoolID).
		Where("id = ?", programID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.StudyProgram{}, err
	}

	item, err := scanProgram(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.StudyProgram{}, def.ErrNotFound
	}
	if err != nil {
		return model.StudyProgram{}, def.NormalizeError(err)
	}
	return item, nil
}

func (r *repository) List(ctx context.Context, schoolID string, filter model.StudyProgramFilter) ([]model.StudyProgram, model.PageMeta, error) {
	countBuilder := def.Builder().
		Select("count(*)").
		From("study_programs").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	selectBuilder := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"name",
			"transmission",
			"total_hours",
			"created_at",
			"updated_at",
		).
		From("study_programs").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")

	if filter.Search != "" {
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
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

	items := make([]model.StudyProgram, 0, limit)
	for rows.Next() {
		item, err := scanProgram(rows)
		if err != nil {
			return nil, model.PageMeta{}, def.NormalizeError(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, model.PageMeta{}, def.NormalizeError(err)
	}

	return items, model.PageMeta{Limit: limit, Offset: offset, Count: total}, nil
}

func (r *repository) Create(ctx context.Context, schoolID string, input model.StudyProgramCreate) (model.StudyProgram, error) {
	query, args, err := def.Builder().
		Insert("study_programs").
		Columns("school_id", "name", "transmission", "total_hours").
		Values(schoolID, input.Name, input.Transmission, input.TotalHours).
		Suffix("RETURNING id::text").
		ToSql()
	if err != nil {
		return model.StudyProgram{}, err
	}

	var programID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&programID); err != nil {
		return model.StudyProgram{}, def.NormalizeError(err)
	}
	return r.GetByID(ctx, schoolID, programID)
}

func (r *repository) Update(ctx context.Context, schoolID, programID string, patch model.StudyProgramPatch) (model.StudyProgram, error) {
	builder := def.Builder().
		Update("study_programs").
		Set("updated_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", programID).
		Where("deleted_at IS NULL")

	if patch.Name != nil {
		builder = builder.Set("name", *patch.Name)
	}
	if patch.Transmission != nil {
		builder = builder.Set("transmission", *patch.Transmission)
	}
	if patch.TotalHours != nil {
		builder = builder.Set("total_hours", *patch.TotalHours)
	}

	query, args, err := builder.Suffix("RETURNING id::text").ToSql()
	if err != nil {
		return model.StudyProgram{}, err
	}

	var updatedID string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.StudyProgram{}, def.ErrNotFound
		}
		return model.StudyProgram{}, def.NormalizeError(err)
	}

	return r.GetByID(ctx, schoolID, updatedID)
}

func (r *repository) Delete(ctx context.Context, schoolID, programID string) error {
	query, args, err := def.Builder().
		Update("study_programs").
		Set("deleted_at", sq.Expr("now()")).
		Set("updated_at", sq.Expr("now()")).
		Where("school_id = ?", schoolID).
		Where("id = ?", programID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return def.ErrNotFound
	}
	return nil
}

func scanProgram(row pgx.Row) (model.StudyProgram, error) {
	var item model.StudyProgram
	err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.Name,
		&item.Transmission,
		&item.TotalHours,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

var _ def.StudyProgramRepository = (*repository)(nil)
