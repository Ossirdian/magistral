package v1

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/iam/internal/converter"
	"github.com/ossirdian/magistral/iam/internal/model"
	repo "github.com/ossirdian/magistral/iam/internal/repository"
	svc "github.com/ossirdian/magistral/iam/internal/service"
	iamv1 "github.com/ossirdian/magistral/shared/pkg/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type api struct {
	iamv1.UnimplementedIAMServiceServer
	service svc.IAMService
}

func New(service svc.IAMService) (*api, error) {
	if service == nil {
		return nil, errors.New("iam service is nil")
	}

	return &api{service: service}, nil
}

func (a *api) CreateUser(ctx context.Context, req *iamv1.CreateUserRequest) (*iamv1.User, error) {
	user, err := a.service.CreateUser(ctx, model.UserCreate{
		SchoolID: req.SchoolId,
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     converter.RoleFromProto(req.GetRole()),
		Status:   converter.StatusFromProto(req.GetStatus()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return converter.UserToProto(user), nil
}

func (a *api) GetUser(ctx context.Context, req *iamv1.GetUserRequest) (*iamv1.User, error) {
	user, err := a.service.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}
	return converter.UserToProto(user), nil
}

func (a *api) GetUserByEmail(ctx context.Context, req *iamv1.GetUserByEmailRequest) (*iamv1.User, error) {
	user, err := a.service.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, mapError(err)
	}
	return converter.UserToProto(user), nil
}

func (a *api) ListUsers(ctx context.Context, req *iamv1.ListUsersRequest) (*iamv1.ListUsersResponse, error) {
	users, meta, err := a.service.ListUsers(ctx, model.UserFilter{
		SchoolID: req.SchoolId,
		Role:     rolePtr(req.Role),
		Status:   statusPtr(req.Status),
		Search:   req.Search,
		Limit:    int(req.GetLimit()),
		Offset:   int(req.GetOffset()),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &iamv1.ListUsersResponse{
		Users: converter.UsersToProto(users),
		Meta: &iamv1.PageMeta{
			Limit:  int32(meta.Limit),
			Offset: int32(meta.Offset),
			Count:  int32(meta.Count),
		},
	}, nil
}

func (a *api) UpdateUser(ctx context.Context, req *iamv1.UpdateUserRequest) (*iamv1.User, error) {
	user, err := a.service.UpdateUser(ctx, req.GetUserId(), model.UserPatch{
		Email:    req.Email,
		Role:     rolePtr(req.Role),
		Status:   statusPtr(req.Status),
		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return converter.UserToProto(user), nil
}

func (a *api) ActivateUser(ctx context.Context, req *iamv1.ActivateUserRequest) (*iamv1.User, error) {
	user, err := a.service.ActivateUser(ctx, req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}
	return converter.UserToProto(user), nil
}

func (a *api) DeactivateUser(ctx context.Context, req *iamv1.DeactivateUserRequest) (*iamv1.User, error) {
	user, err := a.service.DeactivateUser(ctx, req.GetUserId(), req.Reason)
	if err != nil {
		return nil, mapError(err)
	}
	return converter.UserToProto(user), nil
}

func (a *api) SetUserPassword(ctx context.Context, req *iamv1.SetUserPasswordRequest) (*emptypb.Empty, error) {
	if err := a.service.SetUserPassword(ctx, req.GetUserId(), req.GetPassword()); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (a *api) Login(ctx context.Context, req *iamv1.LoginRequest) (*iamv1.LoginResponse, error) {
	session, err := a.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &iamv1.LoginResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    converter.TimeToProto(session.ExpiresAt),
		Identity:     converter.IdentityToProto(session.Identity),
	}, nil
}

func (a *api) RefreshSession(ctx context.Context, req *iamv1.RefreshSessionRequest) (*iamv1.RefreshSessionResponse, error) {
	session, err := a.service.RefreshSession(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, mapError(err)
	}

	return &iamv1.RefreshSessionResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    converter.TimeToProto(session.ExpiresAt),
		Identity:     converter.IdentityToProto(session.Identity),
	}, nil
}

func (a *api) RequestPasswordReset(ctx context.Context, req *iamv1.RequestPasswordResetRequest) (*emptypb.Empty, error) {
	if err := a.service.RequestPasswordReset(ctx, req.GetEmail()); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (a *api) ConfirmPasswordReset(ctx context.Context, req *iamv1.ConfirmPasswordResetRequest) (*emptypb.Empty, error) {
	if err := a.service.ConfirmPasswordReset(ctx, req.GetToken(), req.GetNewPassword()); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (a *api) ValidateToken(ctx context.Context, req *iamv1.ValidateTokenRequest) (*iamv1.ValidateTokenResponse, error) {
	result, err := a.service.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, mapError(err)
	}

	return &iamv1.ValidateTokenResponse{
		UserId:   result.Identity.UserID,
		Role:     converter.RoleToProto(result.Identity.Role),
		SchoolId: result.Identity.SchoolID,
		Status:   converter.StatusToProto(result.Status),
		IsActive: result.IsActive,
		IsValid:  result.IsValid,
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, svc.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, svc.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, repo.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, repo.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func rolePtr(role *iamv1.UserRole) *model.UserRole {
	if role == nil {
		return nil
	}
	value := converter.RoleFromProto(*role)
	return &value
}

func statusPtr(status *iamv1.UserStatus) *model.UserStatus {
	if status == nil {
		return nil
	}
	value := converter.StatusFromProto(*status)
	return &value
}

var _ iamv1.IAMServiceServer = (*api)(nil)
