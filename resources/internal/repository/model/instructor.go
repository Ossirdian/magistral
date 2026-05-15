package model

import "time"

type Instructor struct {
	ID                 string
	SchoolID           string
	BranchID           string
	FullName           string
	Phone              *string
	Email              *string
	Status             InstructorStatus
	AvailabilityStatus AvailabilityStatus
	CurrentCarID       *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type InstructorCreate struct {
	BranchID string
	FullName string
	Phone    *string
	Email    *string
	Status   *InstructorStatus
}

type InstructorPatch struct {
	BranchID *string
	FullName *string
	Phone    Field[string]
	Email    Field[string]
	Status   *InstructorStatus
}

type InstructorFilter struct {
	BranchID           *string
	Status             *InstructorStatus
	AvailabilityStatus *AvailabilityStatus
	Limit              int
	Offset             int
}
