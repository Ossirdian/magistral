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
} = (*Repository)(nil)
