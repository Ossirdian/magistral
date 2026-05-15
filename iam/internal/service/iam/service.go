package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ossirdian/magistral/iam/internal/model"
	repo "github.com/ossirdian/magistral/iam/internal/repository"
	svc "github.com/ossirdian/magistral/iam/internal/service"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type AuthConfig interface {
	JWTSecret() string
	Issuer() string
	AccessTokenTTL() time.Duration
	RefreshTokenTTL() time.Duration
	ResetTokenTTL() time.Duration
}

type Service struct {
	repo   repo.Repository
	auth   AuthConfig
	logger logger.Logger
	now    func() time.Time
}

type accessClaims struct {
	UserID   string           `json:"user_id"`
	Role     model.UserRole   `json:"role"`
	SchoolID *string          `json:"school_id,omitempty"`
	Status   model.UserStatus `json:"status"`
	IsActive bool             `json:"is_active"`
	jwt.RegisteredClaims
}

func NewService(repository repo.Repository, auth AuthConfig, l logger.Logger) *Service {
	if l == nil {
		l = logger.Global()
	}

	return &Service{repo: repository, auth: auth, logger: l, now: time.Now}
}

func (s *Service) CreateUser(ctx context.Context, input model.UserCreate) (model.User, error) {
	input.Email = normalizeEmail(input.Email)
	if err := validateUserCreate(input); err != nil {
		return model.User{}, err
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return model.User{}, err
	}

	return s.repo.CreateUser(ctx, input, passwordHash, s.now().UTC())
}

func (s *Service) GetUser(ctx context.Context, userID string) (model.User, error) {
	if strings.TrimSpace(userID) == "" {
		return model.User{}, svc.ErrInvalidArgument
	}
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return model.User{}, err
	}
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *Service) ListUsers(ctx context.Context, filter model.UserFilter) ([]model.User, model.PageMeta, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListUsers(ctx, filter)
}

func (s *Service) UpdateUser(ctx context.Context, userID string, patch model.UserPatch) (model.User, error) {
	if strings.TrimSpace(userID) == "" {
		return model.User{}, svc.ErrInvalidArgument
	}

	current, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return model.User{}, err
	}

	if patch.Email != nil {
		normalized := normalizeEmail(*patch.Email)
		if err := validateEmail(normalized); err != nil {
			return model.User{}, err
		}
		patch.Email = &normalized
	}

	if err := validateUserPatch(current, patch); err != nil {
		return model.User{}, err
	}

	return s.repo.UpdateUser(ctx, userID, patch)
}

func (s *Service) ActivateUser(ctx context.Context, userID string) (model.User, error) {
	active := true
	status := model.UserStatusActive
	return s.UpdateUser(ctx, userID, model.UserPatch{Status: &status, IsActive: &active})
}

func (s *Service) DeactivateUser(ctx context.Context, userID string, reason *string) (model.User, error) {
	if reason != nil && strings.TrimSpace(*reason) != "" {
		s.logger.WithContext(ctx).Info("deactivating user", "user_id", userID, "reason", strings.TrimSpace(*reason))
	}

	active := false
	status := model.UserStatusInactive
	return s.UpdateUser(ctx, userID, model.UserPatch{Status: &status, IsActive: &active})
}

func (s *Service) SetUserPassword(ctx context.Context, userID, password string) error {
	if strings.TrimSpace(userID) == "" {
		return svc.ErrInvalidArgument
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	return s.repo.SetPasswordHash(ctx, userID, passwordHash, s.now().UTC())
}

func (s *Service) Login(ctx context.Context, email, password string) (model.Session, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return model.Session{}, err
	}
	if strings.TrimSpace(password) == "" {
		return model.Session{}, svc.ErrInvalidArgument
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.Session{}, svc.ErrUnauthorized
		}
		return model.Session{}, err
	}
	if !canAuthenticate(user) {
		return model.Session{}, svc.ErrUnauthorized
	}

	credentials, err := s.repo.GetCredentials(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.Session{}, svc.ErrUnauthorized
		}
		return model.Session{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(password)); err != nil {
		return model.Session{}, svc.ErrUnauthorized
	}

	return s.issueSession(ctx, user)
}

func (s *Service) RefreshSession(ctx context.Context, refreshToken string) (model.Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return model.Session{}, svc.ErrInvalidArgument
	}

	now := s.now().UTC()
	storedToken, err := s.repo.GetRefreshToken(ctx, hashToken(refreshToken))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.Session{}, svc.ErrUnauthorized
		}
		return model.Session{}, err
	}
	if storedToken.RevokedAt != nil || !storedToken.ExpiresAt.After(now) {
		return model.Session{}, svc.ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return model.Session{}, err
	}
	if !canAuthenticate(user) {
		return model.Session{}, svc.ErrUnauthorized
	}

	accessToken, accessExpiry, err := s.generateAccessToken(user)
	if err != nil {
		return model.Session{}, err
	}

	newRefreshRaw := generateOpaqueToken()
	refreshed, err := s.repo.RotateRefreshToken(ctx, storedToken.ID, hashToken(newRefreshRaw), now.Add(s.auth.RefreshTokenTTL()), now)
	if err != nil {
		return model.Session{}, err
	}

	return model.Session{
		AccessToken:  accessToken,
		RefreshToken: newRefreshRaw,
		ExpiresAt:    accessExpiry,
		Identity:     model.Identity{UserID: refreshed.UserID, Role: user.Role, SchoolID: user.SchoolID},
	}, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return err
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil
		}
		return err
	}

	now := s.now().UTC()
	if err := s.repo.CreatePasswordResetToken(ctx, user.ID, hashToken(generateOpaqueToken()), now.Add(s.auth.ResetTokenTTL()), now); err != nil {
		return err
	}

	s.logger.WithContext(ctx).Info("password reset token created", "user_id", user.ID)
	return nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if strings.TrimSpace(token) == "" {
		return svc.ErrInvalidArgument
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.repo.ConsumePasswordResetToken(ctx, hashToken(token), passwordHash, s.now().UTC()); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return svc.ErrUnauthorized
		}
		return err
	}

	return nil
}

func (s *Service) ValidateToken(ctx context.Context, accessToken string) (model.TokenValidation, error) {
	if strings.TrimSpace(accessToken) == "" {
		return model.TokenValidation{IsValid: false}, nil
	}

	claims, err := s.parseAccessToken(accessToken)
	if err != nil {
		return model.TokenValidation{IsValid: false}, nil
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return model.TokenValidation{IsValid: false}, nil
	}
	if !canAuthenticate(user) {
		return model.TokenValidation{IsValid: false}, nil
	}

	return model.TokenValidation{
		Identity: model.Identity{UserID: claims.UserID, Role: claims.Role, SchoolID: claims.SchoolID},
		Status:   claims.Status,
		IsActive: claims.IsActive,
		IsValid:  true,
	}, nil
}

func (s *Service) issueSession(ctx context.Context, user model.User) (model.Session, error) {
	now := s.now().UTC()
	accessToken, accessExpiry, err := s.generateAccessToken(user)
	if err != nil {
		return model.Session{}, err
	}

	refreshRaw := generateOpaqueToken()
	refreshToken, err := s.repo.CreateRefreshToken(ctx, user.ID, hashToken(refreshRaw), now.Add(s.auth.RefreshTokenTTL()), now)
	if err != nil {
		return model.Session{}, err
	}

	return model.Session{
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
		ExpiresAt:    accessExpiry,
		Identity:     model.Identity{UserID: refreshToken.UserID, Role: user.Role, SchoolID: user.SchoolID},
	}, nil
}

func (s *Service) generateAccessToken(user model.User) (string, time.Time, error) {
	now := s.now().UTC()
	expiresAt := now.Add(s.auth.AccessTokenTTL())
	claims := accessClaims{
		UserID:   user.ID,
		Role:     user.Role,
		SchoolID: user.SchoolID,
		Status:   user.Status,
		IsActive: user.IsActive,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.auth.Issuer(),
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.auth.JWTSecret()))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *Service) parseAccessToken(accessToken string) (*accessClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &accessClaims{}, func(*jwt.Token) (interface{}, error) {
		return []byte(s.auth.JWTSecret()), nil
	}, jwt.WithIssuer(s.auth.Issuer()))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, svc.ErrUnauthorized
	}

	return claims, nil
}

func validateUserCreate(input model.UserCreate) error {
	if err := validateEmail(input.Email); err != nil {
		return err
	}
	if err := validatePassword(input.Password); err != nil {
		return err
	}
	if err := validateRole(input.Role); err != nil {
		return err
	}
	if err := validateStatus(input.Status); err != nil {
		return err
	}
	if input.Role == model.UserRoleSuperAdmin {
		if input.SchoolID != nil && strings.TrimSpace(*input.SchoolID) != "" {
			return svc.ErrInvalidArgument
		}
	} else if input.SchoolID == nil || strings.TrimSpace(*input.SchoolID) == "" {
		return svc.ErrInvalidArgument
	}
	return nil
}

func validateUserPatch(current model.User, patch model.UserPatch) error {
	if patch.Role != nil {
		if err := validateRole(*patch.Role); err != nil {
			return err
		}
		if *patch.Role == model.UserRoleSuperAdmin && current.SchoolID != nil {
			return svc.ErrInvalidArgument
		}
		if *patch.Role != model.UserRoleSuperAdmin && current.SchoolID == nil {
			return svc.ErrInvalidArgument
		}
	}
	if patch.Status != nil {
		if err := validateStatus(*patch.Status); err != nil {
			return err
		}
	}
	return nil
}

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return svc.ErrInvalidArgument
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return svc.ErrInvalidArgument
	}

	hasLetter, hasDigit := false, false
	for _, r := range password {
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return svc.ErrInvalidArgument
	}
	return nil
}

func validateRole(role model.UserRole) error {
	switch role {
	case model.UserRoleSuperAdmin, model.UserRoleAdmin, model.UserRoleOwner, model.UserRoleInstructor:
		return nil
	default:
		return svc.ErrInvalidArgument
	}
}

func validateStatus(status model.UserStatus) error {
	switch status {
	case model.UserStatusActive, model.UserStatusInactive, model.UserStatusSuspended:
		return nil
	default:
		return svc.ErrInvalidArgument
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateOpaqueToken() string {
	return uuid.NewString() + uuid.NewString()
}

func canAuthenticate(user model.User) bool {
	return user.IsActive && user.Status == model.UserStatusActive
}
