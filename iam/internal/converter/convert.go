package converter

import (
	"time"

	"github.com/ossirdian/magistral/iam/internal/model"
	iamv1 "github.com/ossirdian/magistral/shared/pkg/iam/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserToProto(user model.User) *iamv1.User {
	return &iamv1.User{
		Id:        user.ID,
		SchoolId:  user.SchoolID,
		Email:     user.Email,
		Role:      RoleToProto(user.Role),
		Status:    StatusToProto(user.Status),
		IsActive:  user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func UsersToProto(users []model.User) []*iamv1.User {
	result := make([]*iamv1.User, 0, len(users))
	for _, user := range users {
		result = append(result, UserToProto(user))
	}
	return result
}

func RoleFromProto(role iamv1.UserRole) model.UserRole {
	switch role {
	case iamv1.UserRole_USER_ROLE_SUPER_ADMIN:
		return model.UserRoleSuperAdmin
	case iamv1.UserRole_USER_ROLE_ADMIN:
		return model.UserRoleAdmin
	case iamv1.UserRole_USER_ROLE_OWNER:
		return model.UserRoleOwner
	case iamv1.UserRole_USER_ROLE_INSTRUCTOR:
		return model.UserRoleInstructor
	default:
		return ""
	}
}

func RoleToProto(role model.UserRole) iamv1.UserRole {
	switch role {
	case model.UserRoleSuperAdmin:
		return iamv1.UserRole_USER_ROLE_SUPER_ADMIN
	case model.UserRoleAdmin:
		return iamv1.UserRole_USER_ROLE_ADMIN
	case model.UserRoleOwner:
		return iamv1.UserRole_USER_ROLE_OWNER
	case model.UserRoleInstructor:
		return iamv1.UserRole_USER_ROLE_INSTRUCTOR
	default:
		return iamv1.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func StatusFromProto(status iamv1.UserStatus) model.UserStatus {
	switch status {
	case iamv1.UserStatus_USER_STATUS_ACTIVE:
		return model.UserStatusActive
	case iamv1.UserStatus_USER_STATUS_INACTIVE:
		return model.UserStatusInactive
	case iamv1.UserStatus_USER_STATUS_SUSPENDED:
		return model.UserStatusSuspended
	default:
		return ""
	}
}

func StatusToProto(status model.UserStatus) iamv1.UserStatus {
	switch status {
	case model.UserStatusActive:
		return iamv1.UserStatus_USER_STATUS_ACTIVE
	case model.UserStatusInactive:
		return iamv1.UserStatus_USER_STATUS_INACTIVE
	case model.UserStatusSuspended:
		return iamv1.UserStatus_USER_STATUS_SUSPENDED
	default:
		return iamv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func IdentityToProto(identity model.Identity) *iamv1.Identity {
	return &iamv1.Identity{
		UserId:   identity.UserID,
		Role:     RoleToProto(identity.Role),
		SchoolId: identity.SchoolID,
	}
}

func TimeToProto(value time.Time) *timestamppb.Timestamp {
	return timestamppb.New(value)
}
