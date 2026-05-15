package model

import "time"

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
