package payment

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/reporting/internal/model"
	"github.com/shopspring/decimal"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) PaymentsTotal(ctx context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
	return r.sumDecimal(ctx, "payments", "amount", schoolID, filter)
}

func (r *Repository) RemainingTotal(ctx context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
	where, args := contractsWhere(schoolID, filter)
	return r.queryDecimal(ctx, fmt.Sprintf(`
		SELECT COALESCE(SUM(remaining_amount), 0)::text
		FROM student_contracts
		WHERE %s
	`, where), args...)
}

func (r *Repository) PaymentTypeBreakdown(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.PaymentTypeAmount, error) {
	where, args := paymentsWhere(schoolID, filter)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT type, COALESCE(SUM(amount), 0)::text
		FROM payments
		WHERE %s
		GROUP BY type
		ORDER BY type
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.PaymentTypeAmount, 0)
	for rows.Next() {
		var item model.PaymentTypeAmount
		var amount string
		if err := rows.Scan(&item.Type, &amount); err != nil {
			return nil, err
		}
		item.Amount, err = decimal.NewFromString(amount)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) sumDecimal(ctx context.Context, table, field, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
	where, args := paymentsWhere(schoolID, filter)
	return r.queryDecimal(ctx, fmt.Sprintf(`
		SELECT COALESCE(SUM(%s), 0)::text
		FROM %s
		WHERE %s
	`, field, table, where), args...)
}

func (r *Repository) queryDecimal(ctx context.Context, query string, args ...any) (decimal.Decimal, error) {
	var value string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&value); err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromString(value)
}

func paymentsWhere(schoolID string, filter model.ReportFilter) (string, []any) {
	clauses := []string{"school_id = $1", "deleted_at IS NULL"}
	args := []any{schoolID}
	clauses, args = appendStudentIDsClause(clauses, args, filter)
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

func contractsWhere(schoolID string, filter model.ReportFilter) (string, []any) {
	clauses := []string{"school_id = $1"}
	args := []any{schoolID}
	clauses, args = appendStudentIDsClause(clauses, args, filter)
	return strings.Join(clauses, " AND "), args
}

func appendStudentIDsClause(clauses []string, args []any, filter model.ReportFilter) ([]string, []any) {
	if !filter.StudentIDsFiltered {
		return clauses, args
	}
	if len(filter.StudentIDs) == 0 {
		return append(clauses, "FALSE"), args
	}
	placeholders := make([]string, 0, len(filter.StudentIDs))
	for _, id := range filter.StudentIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
		args = append(args, id)
	}
	return append(clauses, fmt.Sprintf("student_id IN (%s)", strings.Join(placeholders, ", "))), args
}

var _ interface {
	PaymentsTotal(context.Context, string, model.ReportFilter) (decimal.Decimal, error)
	RemainingTotal(context.Context, string, model.ReportFilter) (decimal.Decimal, error)
	PaymentTypeBreakdown(context.Context, string, model.ReportFilter) ([]model.PaymentTypeAmount, error)
} = (*Repository)(nil)
