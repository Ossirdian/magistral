package lesson

import (
	"context"
	"fmt"

	"github.com/ossirdian/magistral/lesson/internal/model"
	def "github.com/ossirdian/magistral/lesson/internal/repository"
)

func (r *repository) List(ctx context.Context, schoolID string, filter model.LessonFilter) ([]model.Lesson, model.PageMeta, error) {
	countBuilder := def.Builder().
		Select("count(*)").
		From("lessons").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	selectBuilder := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"COALESCE(branch_id::text, '')",
			"student_id::text",
			"instructor_id::text",
			"car_id::text",
			"start_at",
			"end_at",
			"status",
			"comment",
			"is_generated",
			"created_at",
			"updated_at",
		).
		From("lessons").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")

	if filter.BranchID != nil {
		countBuilder = countBuilder.Where("branch_id = ?", *filter.BranchID)
		selectBuilder = selectBuilder.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.StudentID != nil {
		countBuilder = countBuilder.Where("student_id = ?", *filter.StudentID)
		selectBuilder = selectBuilder.Where("student_id = ?", *filter.StudentID)
	}
	if filter.InstructorID != nil {
		countBuilder = countBuilder.Where("instructor_id = ?", *filter.InstructorID)
		selectBuilder = selectBuilder.Where("instructor_id = ?", *filter.InstructorID)
	}
	if filter.CarID != nil {
		countBuilder = countBuilder.Where("car_id = ?", *filter.CarID)
		selectBuilder = selectBuilder.Where("car_id = ?", *filter.CarID)
	}
	if filter.Status != nil {
		countBuilder = countBuilder.Where("status = ?", *filter.Status)
		selectBuilder = selectBuilder.Where("status = ?", *filter.Status)
	}
	if filter.DateFrom != nil {
		countBuilder = countBuilder.Where("end_at >= ?", *filter.DateFrom)
		selectBuilder = selectBuilder.Where("end_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		countBuilder = countBuilder.Where("start_at <= ?", *filter.DateTo)
		selectBuilder = selectBuilder.Where("start_at <= ?", *filter.DateTo)
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
		OrderBy(lessonOrderBy(filter)...).
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

	items := make([]model.Lesson, 0, limit)
	for rows.Next() {
		item, scanErr := scanLesson(rows)
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

func lessonOrderBy(filter model.LessonFilter) []string {
	column := "start_at"
	switch filter.SortBy {
	case model.LessonSortByStudentID:
		column = "student_id"
	case model.LessonSortByInstructorID:
		column = "instructor_id"
	case model.LessonSortByCarID:
		column = "car_id"
	case model.LessonSortByStartAt, "":
		column = "start_at"
	}

	direction := "ASC"
	if filter.SortOrder == model.SortOrderDesc {
		direction = "DESC"
	}

	return []string{
		fmt.Sprintf("%s %s", column, direction),
		"start_at ASC",
		"id ASC",
	}
}
