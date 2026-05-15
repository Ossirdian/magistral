package model

import "time"

type User struct {
	ID        string
	SchoolID  *string
	Email     string
	Role      UserRole
	Status    UserStatus
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCreate struct {
	SchoolID *string
	Email    string
	Password string
	Role     UserRole
	Status   UserStatus
}

type UserPatch struct {
	Email    *string
	Role     *UserRole
	Status   *UserStatus
	IsActive *bool
}

type UserFilter struct {
	SchoolID *string
	Role     *UserRole
	Status   *UserStatus
	Search   *string
	Limit    int
	Offset   int
}

type Credentials struct {
	UserID            string
	PasswordHash      string
	PasswordUpdatedAt time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Identity struct {
	UserID   string
	Role     UserRole
	SchoolID *string
}

type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Identity     Identity
}

type TokenValidation struct {
	Identity Identity
	Status   UserStatus
	IsActive bool
	IsValid  bool
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleOwner      UserRole = "owner"
	UserRoleInstructor UserRole = "instructor"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)
