package repository

import (
	"context"
	"time"

	"github.com/ossirdian/magistral/iam/internal/model"
)

type Repository interface {
	CreateUser(ctx context.Context, input model.UserCreate, passwordHash string, passwordUpdatedAt time.Time) (model.User, error)
	GetUserByID(ctx context.Context, userID string) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	ListUsers(ctx context.Context, filter model.UserFilter) ([]model.User, model.PageMeta, error)
	UpdateUser(ctx context.Context, userID string, patch model.UserPatch) (model.User, error)
	SetPasswordHash(ctx context.Context, userID, passwordHash string, passwordUpdatedAt time.Time) error
	GetCredentials(ctx context.Context, userID string) (model.Credentials, error)
	CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt, createdAt time.Time) error
	ConsumePasswordResetToken(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error
	CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt, createdAt time.Time) (model.RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, currentTokenID, newTokenHash string, expiresAt, now time.Time) (model.RefreshToken, error)
}
