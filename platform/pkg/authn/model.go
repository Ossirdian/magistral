package authn

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleOwner      Role = "owner"
	RoleInstructor Role = "instructor"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type Identity struct {
	UserID   string
	Role     Role
	SchoolID *string
}

type TokenValidation struct {
	Identity Identity
	Status   UserStatus
	IsActive bool
	IsValid  bool
}
