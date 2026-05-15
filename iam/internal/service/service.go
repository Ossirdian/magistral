package service

import (
	"context"

	"github.com/ossirdian/magistral/iam/internal/model"
)

type IAMService interface {
	CreateUser(ctx context.Context, input model.UserCreate) (model.User, error)
	GetUser(ctx context.Context, userID string) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	ListUsers(ctx context.Context, filter model.UserFilter) ([]model.User, model.PageMeta, error)
	UpdateUser(ctx context.Context, userID string, patch model.UserPatch) (model.User, error)
	ActivateUser(ctx context.Context, userID string) (model.User, error)
	DeactivateUser(ctx context.Context, userID string, reason *string) (model.User, error)
	SetUserPassword(ctx context.Context, userID, password string) error
	Login(ctx context.Context, email, password string) (model.Session, error)
	RefreshSession(ctx context.Context, refreshToken string) (model.Session, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
	ValidateToken(ctx context.Context, accessToken string) (model.TokenValidation, error)
}
