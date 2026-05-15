package model

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
