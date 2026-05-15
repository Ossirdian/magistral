package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/payment/internal/model"
	def "github.com/ossirdian/magistral/payment/internal/repository"
	"github.com/shopspring/decimal"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}

func (r *repository) GetByID(ctx context.Context, schoolID, paymentID string) (model.Payment, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"student_id::text",
			"type",
			"amount::text",
			"comment",
			"created_at",
			"updated_at",
		).
		From("payments").
		Where("school_id = ?", schoolID).
		Where("id = ?", paymentID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Payment{}, err
	}

	row := r.pool.QueryRow(ctx, query, args...)
	item, err := scanPayment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, def.ErrNotFound
		}
		return model.Payment{}, def.NormalizeError(err)
	}

	return item, nil
}

func (r *repository) List(ctx context.Context, schoolID string, filter model.PaymentFilter) ([]model.Payment, model.PageMeta, error) {
	countBuilder := def.Builder().
		Select("count(*)").
		From("payments").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")
	selectBuilder := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"student_id::text",
			"type",
			"amount::text",
			"comment",
			"created_at",
			"updated_at",
		).
		From("payments").
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL")

	if filter.StudentID != nil {
		countBuilder = countBuilder.Where("student_id = ?", *filter.StudentID)
		selectBuilder = selectBuilder.Where("student_id = ?", *filter.StudentID)
	}
	if filter.Type != nil {
		countBuilder = countBuilder.Where("type = ?", *filter.Type)
		selectBuilder = selectBuilder.Where("type = ?", *filter.Type)
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

	items := make([]model.Payment, 0, limit)
	for rows.Next() {
		item, scanErr := scanPayment(rows)
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

func (r *repository) Create(ctx context.Context, schoolID string, input model.PaymentCreate) (model.Payment, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Payment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query, args, err := def.Builder().
		Insert("payments").
		Columns("school_id", "student_id", "type", "amount", "comment").
		Values(schoolID, input.StudentID, input.Type, input.Amount.StringFixed(2), input.Comment).
		Suffix("RETURNING id::text").
		ToSql()
	if err != nil {
		return model.Payment{}, err
	}

	var paymentID string
	if err := tx.QueryRow(ctx, query, args...).Scan(&paymentID); err != nil {
		return model.Payment{}, def.NormalizeError(err)
	}

	if _, err := r.recalculateBalanceTx(ctx, tx, schoolID, input.StudentID, nil); err != nil {
		return model.Payment{}, err
	}

	item, err := r.getByIDTx(ctx, tx, schoolID, paymentID)
	if err != nil {
		return model.Payment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Payment{}, err
	}

	return item, nil
}

func (r *repository) Update(ctx context.Context, schoolID, paymentID string, patch model.PaymentPatch) (model.Payment, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Payment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := r.getByIDTx(ctx, tx, schoolID, paymentID)
	if err != nil {
		return model.Payment{}, err
	}

	builder := def.Builder().
		Update("payments").
		Set("updated_at", time.Now().UTC()).
		Where("school_id = ?", schoolID).
		Where("id = ?", paymentID).
		Where("deleted_at IS NULL")

	if patch.Type != nil {
		builder = builder.Set("type", *patch.Type)
	}
	if patch.Amount != nil {
		builder = builder.Set("amount", patch.Amount.StringFixed(2))
	}
	if patch.CommentSet {
		builder = builder.Set("comment", patch.Comment)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Payment{}, err
	}
	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return model.Payment{}, def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.Payment{}, def.ErrNotFound
	}

	if _, err := r.recalculateBalanceTx(ctx, tx, schoolID, current.StudentID, nil); err != nil {
		return model.Payment{}, err
	}

	item, err := r.getByIDTx(ctx, tx, schoolID, paymentID)
	if err != nil {
		return model.Payment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Payment{}, err
	}

	return item, nil
}

func (r *repository) Delete(ctx context.Context, schoolID, paymentID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := r.getByIDTx(ctx, tx, schoolID, paymentID)
	if err != nil {
		return err
	}

	query, args, err := def.Builder().
		Update("payments").
		Set("deleted_at", time.Now().UTC()).
		Set("updated_at", time.Now().UTC()).
		Where("school_id = ?", schoolID).
		Where("id = ?", paymentID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return def.ErrNotFound
	}

	if _, err := r.recalculateBalanceTx(ctx, tx, schoolID, current.StudentID, nil); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repository) GetBalance(ctx context.Context, schoolID, studentID string) (model.StudentBalance, error) {
	var contractTotalText string
	var updatedAt time.Time
	err := r.pool.QueryRow(
		ctx,
		`SELECT contract_total::text, updated_at
		   FROM student_contracts
		  WHERE school_id = $1 AND student_id = $2`,
		schoolID, studentID,
	).Scan(&contractTotalText, &updatedAt)

	contractTotal := decimal.Zero
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return model.StudentBalance{}, def.NormalizeError(err)
		}
		updatedAt = time.Now().UTC()
	} else {
		contractTotal, err = decimal.NewFromString(contractTotalText)
		if err != nil {
			return model.StudentBalance{}, fmt.Errorf("parse contract total: %w", err)
		}
	}

	paidAmount, err := r.sumPaid(ctx, r.pool, schoolID, studentID)
	if err != nil {
		return model.StudentBalance{}, err
	}

	return model.StudentBalance{
		StudentID:       studentID,
		ContractTotal:   contractTotal,
		PaidAmount:      paidAmount,
		RemainingAmount: contractTotal.Sub(paidAmount),
		UpdatedAt:       updatedAt,
	}, nil
}

func (r *repository) UpsertContract(ctx context.Context, schoolID, studentID string, contractTotal decimal.Decimal) (model.StudentBalance, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.StudentBalance{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	balance, err := r.recalculateBalanceTx(ctx, tx, schoolID, studentID, &contractTotal)
	if err != nil {
		return model.StudentBalance{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.StudentBalance{}, err
	}
	return balance, nil
}

func (r *repository) getByIDTx(ctx context.Context, tx pgx.Tx, schoolID, paymentID string) (model.Payment, error) {
	query, args, err := def.Builder().
		Select(
			"id::text",
			"school_id::text",
			"student_id::text",
			"type",
			"amount::text",
			"comment",
			"created_at",
			"updated_at",
		).
		From("payments").
		Where("school_id = ?", schoolID).
		Where("id = ?", paymentID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.Payment{}, err
	}

	item, err := scanPayment(tx.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, def.ErrNotFound
		}
		return model.Payment{}, def.NormalizeError(err)
	}
	return item, nil
}

func (r *repository) recalculateBalanceTx(ctx context.Context, tx pgx.Tx, schoolID, studentID string, contractTotal *decimal.Decimal) (model.StudentBalance, error) {
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO student_contracts (student_id, school_id, contract_total, remaining_amount, updated_at)
		 VALUES ($1, $2, 0, 0, now())
		 ON CONFLICT (student_id) DO NOTHING`,
		studentID, schoolID,
	); err != nil {
		return model.StudentBalance{}, def.NormalizeError(err)
	}

	if contractTotal != nil {
		if _, err := tx.Exec(
			ctx,
			`UPDATE student_contracts
			    SET contract_total = $3,
			        updated_at = now()
			  WHERE student_id = $1 AND school_id = $2`,
			studentID, schoolID, contractTotal.StringFixed(2),
		); err != nil {
			return model.StudentBalance{}, def.NormalizeError(err)
		}
	}

	var contractTotalText string
	var updatedAt time.Time
	if err := tx.QueryRow(
		ctx,
		`SELECT contract_total::text, updated_at
		   FROM student_contracts
		  WHERE student_id = $1 AND school_id = $2`,
		studentID, schoolID,
	).Scan(&contractTotalText, &updatedAt); err != nil {
		return model.StudentBalance{}, def.NormalizeError(err)
	}

	currentTotal, err := decimal.NewFromString(contractTotalText)
	if err != nil {
		return model.StudentBalance{}, fmt.Errorf("parse contract total: %w", err)
	}

	paidAmount, err := r.sumPaid(ctx, tx, schoolID, studentID)
	if err != nil {
		return model.StudentBalance{}, err
	}
	remainingAmount := currentTotal.Sub(paidAmount)

	if _, err := tx.Exec(
		ctx,
		`UPDATE student_contracts
		    SET remaining_amount = $3,
		        updated_at = now()
		  WHERE student_id = $1 AND school_id = $2`,
		studentID, schoolID, remainingAmount.StringFixed(2),
	); err != nil {
		return model.StudentBalance{}, def.NormalizeError(err)
	}

	return model.StudentBalance{
		StudentID:       studentID,
		ContractTotal:   currentTotal,
		PaidAmount:      paidAmount,
		RemainingAmount: remainingAmount,
		UpdatedAt:       time.Now().UTC(),
	}, nil
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *repository) sumPaid(ctx context.Context, runner queryRower, schoolID, studentID string) (decimal.Decimal, error) {
	var paidAmountText string
	if err := runner.QueryRow(
		ctx,
		`SELECT COALESCE(sum(amount), 0)::text
		   FROM payments
		  WHERE school_id = $1
		    AND student_id = $2
		    AND deleted_at IS NULL`,
		schoolID, studentID,
	).Scan(&paidAmountText); err != nil {
		return decimal.Zero, def.NormalizeError(err)
	}

	paidAmount, err := decimal.NewFromString(strings.TrimSpace(paidAmountText))
	if err != nil {
		return decimal.Zero, fmt.Errorf("parse paid amount: %w", err)
	}
	return paidAmount, nil
}

type paymentScanner interface {
	Scan(dest ...any) error
}

func scanPayment(scanner paymentScanner) (model.Payment, error) {
	var (
		item        model.Payment
		paymentType string
		amountText  string
		comment     *string
	)

	if err := scanner.Scan(
		&item.ID,
		&item.SchoolID,
		&item.StudentID,
		&paymentType,
		&amountText,
		&comment,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return model.Payment{}, err
	}

	amount, err := decimal.NewFromString(amountText)
	if err != nil {
		return model.Payment{}, fmt.Errorf("parse payment amount: %w", err)
	}

	item.Type = model.PaymentType(paymentType)
	item.Amount = amount
	item.Comment = comment
	return item, nil
}
