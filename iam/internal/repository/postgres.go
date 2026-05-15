package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/magistral/iam/internal/model"
	"github.com/ossirdian/magistral/platform/pkg/logger"
)

type PostgresRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewPostgresRepository(pool *pgxpool.Pool, l logger.Logger) *PostgresRepository {
	if l == nil {
		l = logger.Global()
	}

	return &PostgresRepository{pool: pool, logger: l}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, input model.UserCreate, passwordHash string, passwordUpdatedAt time.Time) (model.User, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := Builder().
		Insert("users").
		Columns("school_id", "email", "role", "status", "is_active").
		Values(input.SchoolID, strings.ToLower(input.Email), string(input.Role), string(input.Status), input.Status == model.UserStatusActive).
		Suffix("RETURNING id, school_id, email, role, status, is_active, created_at, updated_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return model.User{}, err
	}

	user, err := scanUser(tx.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.User{}, NormalizeError(err)
	}

	query = Builder().
		Insert("credentials").
		Columns("user_id", "password_hash", "password_updated_at").
		Values(user.ID, passwordHash, passwordUpdatedAt)

	sqlStr, args, err = query.ToSql()
	if err != nil {
		return model.User{}, err
	}

	if _, err := tx.Exec(ctx, sqlStr, args...); err != nil {
		return model.User{}, NormalizeError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, userID string) (model.User, error) {
	return r.getUser(ctx, Builder().
		Select("id", "school_id", "email", "role", "status", "is_active", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": userID}))
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	return r.getUser(ctx, Builder().
		Select("id", "school_id", "email", "role", "status", "is_active", "created_at", "updated_at").
		From("users").
		Where("LOWER(email) = LOWER(?)", email))
}

func (r *PostgresRepository) ListUsers(ctx context.Context, filter model.UserFilter) ([]model.User, model.PageMeta, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := applyUserFilter(Builder().
		Select("id", "school_id", "email", "role", "status", "is_active", "created_at", "updated_at").
		From("users"), filter).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	countQuery := applyUserFilter(Builder().
		Select("COUNT(*)").
		From("users"), filter)

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, model.PageMeta{}, err
	}

	var count int
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&count); err != nil {
		return nil, model.PageMeta{}, err
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, model.PageMeta{}, err
	}

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, model.PageMeta{}, err
	}
	defer rows.Close()

	users := make([]model.User, 0, limit)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, model.PageMeta{}, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, model.PageMeta{}, err
	}

	return users, model.PageMeta{Limit: limit, Offset: offset, Count: count}, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, userID string, patch model.UserPatch) (model.User, error) {
	update := Builder().
		Update("users").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})

	changed := false
	if patch.Email != nil {
		update = update.Set("email", strings.ToLower(*patch.Email))
		changed = true
	}
	if patch.Role != nil {
		update = update.Set("role", string(*patch.Role))
		changed = true
	}
	if patch.Status != nil {
		update = update.Set("status", string(*patch.Status))
		changed = true
	}
	if patch.IsActive != nil {
		update = update.Set("is_active", *patch.IsActive)
		changed = true
	}
	if !changed {
		return r.GetUserByID(ctx, userID)
	}

	update = update.Suffix("RETURNING id, school_id, email, role, status, is_active, created_at, updated_at")

	sqlStr, args, err := update.ToSql()
	if err != nil {
		return model.User{}, err
	}

	user, err := scanUser(r.pool.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.User{}, NormalizeError(err)
	}

	return user, nil
}

func (r *PostgresRepository) SetPasswordHash(ctx context.Context, userID, passwordHash string, passwordUpdatedAt time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := Builder().
		Insert("credentials").
		Columns("user_id", "password_hash", "password_updated_at").
		Values(userID, passwordHash, passwordUpdatedAt).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, password_updated_at = EXCLUDED.password_updated_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, sqlStr, args...); err != nil {
		return NormalizeError(err)
	}

	updateUser := Builder().
		Update("users").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})

	sqlStr, args, err = updateUser.ToSql()
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, sqlStr, args...)
	if err != nil {
		return NormalizeError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetCredentials(ctx context.Context, userID string) (model.Credentials, error) {
	query := Builder().
		Select("user_id", "password_hash", "password_updated_at").
		From("credentials").
		Where(sq.Eq{"user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return model.Credentials{}, err
	}

	credentials, err := scanCredentials(r.pool.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.Credentials{}, NormalizeError(err)
	}

	return credentials, nil
}

func (r *PostgresRepository) CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt, createdAt time.Time) error {
	query := Builder().
		Insert("password_reset_tokens").
		Columns("user_id", "token_hash", "expires_at", "created_at").
		Values(userID, tokenHash, expiresAt, createdAt)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return NormalizeError(err)
}

func (r *PostgresRepository) ConsumePasswordResetToken(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := Builder().
		Update("password_reset_tokens").
		Set("used_at", now).
		Where(sq.Eq{"token_hash": tokenHash}).
		Where("used_at IS NULL").
		Where("expires_at > ?", now).
		Suffix("RETURNING user_id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	var userID string
	if err := tx.QueryRow(ctx, sqlStr, args...).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return NormalizeError(err)
	}

	upsertCredentials := Builder().
		Insert("credentials").
		Columns("user_id", "password_hash", "password_updated_at").
		Values(userID, newPasswordHash, now).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, password_updated_at = EXCLUDED.password_updated_at")

	sqlStr, args, err = upsertCredentials.ToSql()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, sqlStr, args...); err != nil {
		return NormalizeError(err)
	}

	updateUser := Builder().
		Update("users").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})

	sqlStr, args, err = updateUser.ToSql()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, sqlStr, args...); err != nil {
		return NormalizeError(err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt, createdAt time.Time) (model.RefreshToken, error) {
	query := Builder().
		Insert("refresh_tokens").
		Columns("user_id", "token_hash", "expires_at", "created_at").
		Values(userID, tokenHash, expiresAt, createdAt).
		Suffix("RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return model.RefreshToken{}, err
	}

	token, err := scanRefreshToken(r.pool.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.RefreshToken{}, NormalizeError(err)
	}

	return token, nil
}

func (r *PostgresRepository) GetRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	query := Builder().
		Select("id, user_id, token_hash, expires_at, revoked_at, created_at").
		From("refresh_tokens").
		Where(sq.Eq{"token_hash": tokenHash})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return model.RefreshToken{}, err
	}

	token, err := scanRefreshToken(r.pool.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.RefreshToken{}, NormalizeError(err)
	}

	return token, nil
}

func (r *PostgresRepository) RotateRefreshToken(ctx context.Context, currentTokenID, newTokenHash string, expiresAt, now time.Time) (model.RefreshToken, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.RefreshToken{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := Builder().
		Update("refresh_tokens").
		Set("revoked_at", now).
		Where(sq.Eq{"id": currentTokenID}).
		Where("revoked_at IS NULL").
		Suffix("RETURNING user_id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return model.RefreshToken{}, err
	}

	var userID string
	if err := tx.QueryRow(ctx, sqlStr, args...).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}
		return model.RefreshToken{}, NormalizeError(err)
	}

	insertToken := Builder().
		Insert("refresh_tokens").
		Columns("user_id", "token_hash", "expires_at", "created_at").
		Values(userID, newTokenHash, expiresAt, now).
		Suffix("RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at")

	sqlStr, args, err = insertToken.ToSql()
	if err != nil {
		return model.RefreshToken{}, err
	}

	token, err := scanRefreshToken(tx.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.RefreshToken{}, NormalizeError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.RefreshToken{}, err
	}

	return token, nil
}

func (r *PostgresRepository) getUser(ctx context.Context, query sq.SelectBuilder) (model.User, error) {
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return model.User{}, err
	}

	user, err := scanUser(r.pool.QueryRow(ctx, sqlStr, args...))
	if err != nil {
		return model.User{}, NormalizeError(err)
	}

	return user, nil
}

func applyUserFilter[T interface {
	Where(pred interface{}, args ...interface{}) T
}](query T, filter model.UserFilter) T {
	if filter.SchoolID != nil && *filter.SchoolID != "" {
		query = query.Where(sq.Eq{"school_id": *filter.SchoolID})
	}
	if filter.Role != nil {
		query = query.Where(sq.Eq{"role": string(*filter.Role)})
	}
	if filter.Status != nil {
		query = query.Where(sq.Eq{"status": string(*filter.Status)})
	}
	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(*filter.Search))+"%")
	}
	return query
}

func scanUser(row rowScanner) (model.User, error) {
	var user model.User
	if err := row.Scan(&user.ID, &user.SchoolID, &user.Email, &user.Role, &user.Status, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	return user, nil
}

func scanCredentials(row rowScanner) (model.Credentials, error) {
	var credentials model.Credentials
	if err := row.Scan(&credentials.UserID, &credentials.PasswordHash, &credentials.PasswordUpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Credentials{}, ErrNotFound
		}
		return model.Credentials{}, err
	}
	return credentials, nil
}

func scanRefreshToken(row rowScanner) (model.RefreshToken, error) {
	var token model.RefreshToken
	if err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.RevokedAt, &token.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}
		return model.RefreshToken{}, err
	}
	return token, nil
}
