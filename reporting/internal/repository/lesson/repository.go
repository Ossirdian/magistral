package lesson

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/reporting/internal/model"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CountLessons(ctx context.Context, schoolID string, filter model.ReportFilter) (int, error) {
	where, args := lessonsWhere(schoolID, filter)
	var count int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(1)
		FROM lessons
		WHERE %s
	`, where), args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) LessonsReport(ctx context.Context, schoolID string, filter model.ReportFilter) (model.LessonsReport, error) {
	where, args := lessonsWhere(schoolID, filter)
	var report model.LessonsReport
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			COUNT(1),
			COUNT(1) FILTER (WHERE status = 'scheduled'),
			COUNT(1) FILTER (WHERE status = 'completed'),
			COUNT(1) FILTER (WHERE status = 'cancelled'),
			COUNT(1) FILTER (WHERE status = 'missed')
		FROM lessons
		WHERE %s
	`, where), args...).Scan(
		&report.Total,
		&report.Scheduled,
		&report.Completed,
		&report.Cancelled,
		&report.Missed,
	); err != nil {
		return model.LessonsReport{}, err
	}
	return report, nil
}

func (r *Repository) InstructorLoads(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.EntityLoad, error) {
	return r.entityLoads(ctx, schoolID, filter, "instructor_id")
}

func (r *Repository) CarLoads(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.EntityLoad, error) {
	return r.entityLoads(ctx, schoolID, filter, "car_id")
}

func (r *Repository) entityLoads(ctx context.Context, schoolID string, filter model.ReportFilter, field string) ([]model.EntityLoad, error) {
	where, args := lessonsWhere(schoolID, filter)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT
			%s,
			COUNT(1),
			COALESCE(SUM(EXTRACT(EPOCH FROM (end_at - start_at)) / 60), 0)::bigint
		FROM lessons
		WHERE %s
		GROUP BY %s
		ORDER BY COUNT(1) DESC, %s
	`, field, where, field, field), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.EntityLoad, 0)
	for rows.Next() {
		var item model.EntityLoad
		if err := rows.Scan(&item.ID, &item.LessonsTotal, &item.TotalMinutes); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func lessonsWhere(schoolID string, filter model.ReportFilter) (string, []any) {
	clauses := []string{"school_id = $1", "deleted_at IS NULL"}
	args := []any{schoolID}
	if filter.BranchID != nil && strings.TrimSpace(*filter.BranchID) != "" {
		clauses = append(clauses, fmt.Sprintf("branch_id = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(*filter.BranchID))
	}
	if filter.StudentIDsFiltered {
		if len(filter.StudentIDs) == 0 {
			clauses = append(clauses, "FALSE")
		} else {
			placeholders := make([]string, 0, len(filter.StudentIDs))
			for _, id := range filter.StudentIDs {
				placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
				args = append(args, id)
			}
			clauses = append(clauses, fmt.Sprintf("student_id IN (%s)", strings.Join(placeholders, ", ")))
		}
	}
	if filter.DateFrom != nil {
		clauses = append(clauses, fmt.Sprintf("start_at >= $%d", len(args)+1))
		args = append(args, *filter.DateFrom)
	}
	if filter.DateTo != nil {
		clauses = append(clauses, fmt.Sprintf("start_at < $%d", len(args)+1))
		args = append(args, *filter.DateTo)
	}
	return strings.Join(clauses, " AND "), args
}

var _ interface {
	CountLessons(context.Context, string, model.ReportFilter) (int, error)
	LessonsReport(context.Context, string, model.ReportFilter) (model.LessonsReport, error)
	InstructorLoads(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error)
	CarLoads(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error)
} = (*Repository)(nil)
