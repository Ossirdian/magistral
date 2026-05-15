package car

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func (r *repository) AssignInstructor(ctx context.Context, schoolID, carID string, input model.CarAssignment) (model.CarAssignmentResult, error) {
	repoInput := converter.CarAssignmentFromModel(input)
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.CarAssignmentResult{}, def.NormalizeError(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err = getInstructorForUpdate(ctx, tx, schoolID, repoInput.InstructorID); err != nil {
		return model.CarAssignmentResult{}, err
	}

	previousInstructorID, err := getCarInstructorForUpdate(ctx, tx, schoolID, carID)
	if err != nil {
		return model.CarAssignmentResult{}, err
	}

	appliedAt := time.Now().UTC()
	if repoInput.EffectiveFrom != nil {
		appliedAt = repoInput.EffectiveFrom.UTC()
	}

	query, args, err := def.Builder().
		Update("cars").
		Set("instructor_id", repoInput.InstructorID).
		Set("updated_at", sq.Expr("now()")).
		Where("id = ?", carID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return model.CarAssignmentResult{}, err
	}
	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return model.CarAssignmentResult{}, def.NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.CarAssignmentResult{}, def.ErrNotFound
	}

	if err = tx.Commit(ctx); err != nil {
		return model.CarAssignmentResult{}, def.NormalizeError(err)
	}

	return converter.CarAssignmentResultToModel(repomodel.CarAssignmentResult{
		CarID:                carID,
		InstructorID:         repoInput.InstructorID,
		PreviousInstructorID: previousInstructorID,
		AssignedAt:           appliedAt,
	}), nil
}

func getInstructorForUpdate(ctx context.Context, tx pgx.Tx, schoolID, instructorID string) (string, error) {
	var id string
	query, args, err := def.Builder().
		Select("id::text").
		From("instructors").
		Where("id = ?", instructorID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return "", err
	}
	if err := tx.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", def.ErrNotFound
		}
		return "", def.NormalizeError(err)
	}

	return id, nil
}

func getCarInstructorForUpdate(ctx context.Context, tx pgx.Tx, schoolID, carID string) (*string, error) {
	var instructorID *string
	query, args, err := def.Builder().
		Select("instructor_id::text").
		From("cars").
		Where("id = ?", carID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := tx.QueryRow(ctx, query, args...).Scan(&instructorID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, def.ErrNotFound
		}
		return nil, def.NormalizeError(err)
	}

	return instructorID, nil
}
