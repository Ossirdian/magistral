package car

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/ossirdian/magistral/resources/internal/model"
	def "github.com/ossirdian/magistral/resources/internal/repository"
	"github.com/ossirdian/magistral/resources/internal/repository/converter"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func (r *repository) GetAvailability(ctx context.Context, schoolID, carID string) ([]model.AvailabilityWindow, error) {
	if _, err := r.GetByID(ctx, schoolID, carID); err != nil {
		return nil, err
	}

	items, err := listAvailability(ctx, r.pool, carID)
	if err != nil {
		return nil, err
	}

	return converter.AvailabilityWindowsToModel(items), nil
}

func (r *repository) ReplaceAvailability(ctx context.Context, schoolID, carID string, windows []model.AvailabilityWindowInput) ([]model.AvailabilityWindow, error) {
	repoWindows := converter.AvailabilityWindowInputFromModel(windows)
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, def.NormalizeError(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err = ensureCarExists(ctx, tx, schoolID, carID); err != nil {
		return nil, err
	}

	deleteQuery, deleteArgs, err := def.Builder().
		Delete("car_availability").
		Where("car_id = ?", carID).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, deleteQuery, deleteArgs...); err != nil {
		return nil, def.NormalizeError(err)
	}

	for _, window := range repoWindows {
		insertQuery, insertArgs, buildErr := def.Builder().
			Insert("car_availability").
			Columns("car_id", "type", "start_at", "end_at", "comment").
			Values(carID, window.Type, window.StartAt, window.EndAt, window.Comment).
			ToSql()
		if buildErr != nil {
			return nil, buildErr
		}
		if _, err = tx.Exec(ctx, insertQuery, insertArgs...); err != nil {
			return nil, def.NormalizeError(err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, def.NormalizeError(err)
	}

	items, err := listAvailability(ctx, r.pool, carID)
	if err != nil {
		return nil, err
	}

	return converter.AvailabilityWindowsToModel(items), nil
}

func ensureCarExists(ctx context.Context, tx pgx.Tx, schoolID, carID string) error {
	var stub int
	query, args, err := def.Builder().
		Select("1").
		From("cars").
		Where("id = ?", carID).
		Where("school_id = ?", schoolID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, query, args...).Scan(&stub); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return def.ErrNotFound
		}
		return def.NormalizeError(err)
	}

	return nil
}

func listAvailability(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, carID string,
) ([]repomodel.AvailabilityWindow, error) {
	query, args, err := def.Builder().
		Select("id::text", "type", "start_at", "end_at", "comment").
		From("car_availability").
		Where("car_id = ?", carID).
		OrderBy("start_at ASC", "end_at ASC", "id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, def.NormalizeError(err)
	}
	defer rows.Close()

	items := make([]repomodel.AvailabilityWindow, 0)
	for rows.Next() {
		var item repomodel.AvailabilityWindow
		if err = rows.Scan(&item.ID, &item.Type, &item.StartAt, &item.EndAt, &item.Comment); err != nil {
			return nil, def.NormalizeError(err)
		}
		item.ResourceType = repomodel.ResourceTypeCar
		item.ResourceID = carID
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, def.NormalizeError(err)
	}

	return items, nil
}
