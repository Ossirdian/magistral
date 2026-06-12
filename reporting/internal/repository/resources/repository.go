package resources

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

func (r *Repository) CountInstructors(ctx context.Context, schoolID string, branchID *string) (int, error) {
	where, args := resourceWhere(schoolID, branchID)
	var count int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(1)
		FROM instructors
		WHERE %s
	`, where), args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountCars(ctx context.Context, schoolID string, branchID *string) (int, error) {
	where, args := resourceWhere(schoolID, branchID)
	var count int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(1)
		FROM cars
		WHERE %s
	`, where), args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) ListInstructors(ctx context.Context, schoolID string, branchID *string) ([]model.InstructorRef, error) {
	where, args := resourceWhere(schoolID, branchID)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, full_name
		FROM instructors
		WHERE %s
		ORDER BY full_name, id
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.InstructorRef, 0)
	for rows.Next() {
		var item model.InstructorRef
		if err := rows.Scan(&item.ID, &item.FullName); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) ListCars(ctx context.Context, schoolID string, branchID *string) ([]model.CarRef, error) {
	where, args := resourceWhere(schoolID, branchID)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, plate_number, COALESCE(brand, ''), COALESCE(model, '')
		FROM cars
		WHERE %s
		ORDER BY plate_number, id
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.CarRef, 0)
	for rows.Next() {
		var item model.CarRef
		var brand string
		var modelName string
		if err := rows.Scan(&item.ID, &item.PlateNumber, &brand, &modelName); err != nil {
			return nil, err
		}
		item.DisplayName = strings.TrimSpace(strings.Join([]string{item.PlateNumber, strings.TrimSpace(brand), strings.TrimSpace(modelName)}, " "))
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) InstructorAvailabilityMinutes(ctx context.Context, schoolID string, filter model.ReportFilter) (map[string]int, error) {
	return r.availabilityMinutes(ctx, availabilityQueryConfig{
		resourceTable:     "instructors",
		availabilityTable: "instructor_availability",
		resourceIDColumn:  "instructor_id",
	}, schoolID, filter)
}

func (r *Repository) CarAvailabilityMinutes(ctx context.Context, schoolID string, filter model.ReportFilter) (map[string]int, error) {
	return r.availabilityMinutes(ctx, availabilityQueryConfig{
		resourceTable:     "cars",
		availabilityTable: "car_availability",
		resourceIDColumn:  "car_id",
	}, schoolID, filter)
}

type availabilityQueryConfig struct {
	resourceTable     string
	availabilityTable string
	resourceIDColumn  string
}

func (r *Repository) availabilityMinutes(ctx context.Context, cfg availabilityQueryConfig, schoolID string, filter model.ReportFilter) (map[string]int, error) {
	var dateFrom any
	if filter.DateFrom != nil {
		dateFrom = *filter.DateFrom
	}
	var dateTo any
	if filter.DateTo != nil {
		dateTo = *filter.DateTo
	}
	args := []any{schoolID, dateFrom, dateTo}
	where := []string{"res.school_id = $1", "res.deleted_at IS NULL"}
	if filter.BranchID != nil && strings.TrimSpace(*filter.BranchID) != "" {
		args = append(args, strings.TrimSpace(*filter.BranchID))
		where = append(where, fmt.Sprintf("res.branch_id = $%d", len(args)))
	}

	query := fmt.Sprintf(`
		SELECT res.id::text,
			GREATEST(0, COALESCE(SUM(
				CASE
					WHEN av.type = 'working_hours' THEN 1
					WHEN av.type IN ('vacation', 'blackout') THEN -1
					ELSE 0
				END * EXTRACT(EPOCH FROM (
					LEAST(av.end_at, COALESCE($3::timestamptz, av.end_at)) -
					GREATEST(av.start_at, COALESCE($2::timestamptz, av.start_at))
				)) / 60
			), 0))::bigint AS available_minutes
		FROM %s res
		LEFT JOIN %s av ON av.%s = res.id
			AND av.type IN ('working_hours', 'vacation', 'blackout')
			AND ($2::timestamptz IS NULL OR av.end_at > $2::timestamptz)
			AND ($3::timestamptz IS NULL OR av.start_at < $3::timestamptz)
		WHERE %s
		GROUP BY res.id
	`, cfg.resourceTable, cfg.availabilityTable, cfg.resourceIDColumn, strings.Join(where, " AND "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var id string
		var minutes int
		if err := rows.Scan(&id, &minutes); err != nil {
			return nil, err
		}
		result[id] = minutes
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func resourceWhere(schoolID string, branchID *string) (string, []any) {
	clauses := []string{"school_id = $1", "deleted_at IS NULL"}
	args := []any{schoolID}
	if branchID != nil && strings.TrimSpace(*branchID) != "" {
		clauses = append(clauses, fmt.Sprintf("branch_id = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(*branchID))
	}
	return strings.Join(clauses, " AND "), args
}

var _ interface {
	CountInstructors(context.Context, string, *string) (int, error)
	CountCars(context.Context, string, *string) (int, error)
	ListInstructors(context.Context, string, *string) ([]model.InstructorRef, error)
	ListCars(context.Context, string, *string) ([]model.CarRef, error)
	InstructorAvailabilityMinutes(context.Context, string, model.ReportFilter) (map[string]int, error)
	CarAvailabilityMinutes(context.Context, string, model.ReportFilter) (map[string]int, error)
} = (*Repository)(nil)
