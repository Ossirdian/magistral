package model

import "time"

type School struct {
	ID        string
	Name      string
	Status    SchoolStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type SchoolCreate struct {
	Name   string
	Status *SchoolStatus
}

type SchoolAdminBootstrap struct {
	Email    string
	Password string
}

type SchoolPatch struct {
	Name   *string
	Status *SchoolStatus
}

type SchoolFilter struct {
	Status *SchoolStatus
	Search *string
	Limit  int
	Offset int
}

type Branch struct {
	ID        string
	SchoolID  string
	Name      string
	Address   *string
	Status    BranchStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type BranchCreate struct {
	SchoolID string
	Name     string
	Address  *string
	Status   *BranchStatus
}

type BranchPatch struct {
	SchoolID *string
	Name     *string
	Address  Field[string]
	Status   *BranchStatus
}

type BranchFilter struct {
	SchoolID *string
	Status   *BranchStatus
	Limit    int
	Offset   int
}

type Field[T any] struct {
	Set   bool
	Value *T
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}

type SchoolStatus string

const (
	SchoolStatusActive    SchoolStatus = "active"
	SchoolStatusInactive  SchoolStatus = "inactive"
	SchoolStatusSuspended SchoolStatus = "suspended"
)

type BranchStatus string

const (
	BranchStatusActive   BranchStatus = "active"
	BranchStatusInactive BranchStatus = "inactive"
)
