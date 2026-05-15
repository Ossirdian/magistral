package student

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

func (r *Repository) CountActiveStudents(ctx context.Context, schoolID string, filter model.ReportFilter) (int, error) {
	where, args := studentWhere(schoolID, filter)
	var count int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(1)
		FROM students
		WHERE %s AND status = 'active'
	`, where), args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) StudentReport(ctx context.Context, schoolID string, filter model.ReportFilter) (model.StudentsReport, error) {
	where, args := studentWhere(schoolID, filter)
	var report model.StudentsReport
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			COUNT(1),
			COUNT(1) FILTER (WHERE status = 'active'),
			COUNT(1) FILTER (WHERE status = 'new')
		FROM students
		WHERE %s
	`, where), args...).Scan(&report.Total, &report.Active, &report.New); err != nil {
		return model.StudentsReport{}, err
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT status, COUNT(1)
		FROM students
		WHERE %s
		GROUP BY status
		ORDER BY status
	`, where), args...)
	if err != nil {
		return model.StudentsReport{}, err
	}
	defer rows.Close()

	report.Statuses = make([]model.StatusCount, 0)
	for rows.Next() {
		var item model.StatusCount
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			return model.StudentsReport{}, err
		}
		report.Statuses = append(report.Statuses, item)
	}
	if err := rows.Err(); err != nil {
		return model.StudentsReport{}, err
	}

	return report, nil
}

func (r *Repository) StudentIDs(ctx context.Context, schoolID string, filter model.ReportFilter) ([]string, error) {
	where, args := studentWhere(schoolID, filter)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id
		FROM students
		WHERE %s
		ORDER BY id
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func studentWhere(schoolID string, filter model.ReportFilter) (string, []any) {
	clauses := []string{"school_id = $1", "deleted_at IS NULL"}
	args := []any{schoolID}
	if filter.BranchID != nil && strings.TrimSpace(*filter.BranchID) != "" {
		clauses = append(clauses, fmt.Sprintf("branch_id = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(*filter.BranchID))
	}
	if filter.GroupID != nil && strings.TrimSpace(*filter.GroupID) != "" {
		clauses = append(clauses, fmt.Sprintf("group_id = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(*filter.GroupID))
	}
	if filter.DateFrom != nil {
		clauses = append(clauses, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, *filter.DateFrom)
	}
	if filter.DateTo != nil {
		clauses = append(clauses, fmt.Sprintf("created_at < $%d", len(args)+1))
		args = append(args, *filter.DateTo)
	}
	return strings.Join(clauses, " AND "), args
}

var _ interface {
	CountActiveStudents(context.Context, string, model.ReportFilter) (int, error)
	StudentReport(context.Context, string, model.ReportFilter) (model.StudentsReport, error)
	StudentIDs(context.Context, string, model.ReportFilter) ([]string, error)
} = (*Repository)(nil)
