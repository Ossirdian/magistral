package model

import "time"

type Car struct {
	ID                 string
	SchoolID           string
	BranchID           string
	InstructorID       *string
	PlateNumber        string
	Brand              *string
	Model              *string
	Color              *string
	Transmission       TransmissionType
	Status             CarStatus
	AvailabilityStatus AvailabilityStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CarCreate struct {
	BranchID     string
	InstructorID *string
	PlateNumber  string
	Brand        *string
	Model        *string
	Color        *string
	Transmission *TransmissionType
	Status       *CarStatus
}

type CarPatch struct {
	BranchID     *string
	InstructorID Field[string]
	PlateNumber  *string
	Brand        Field[string]
	Model        Field[string]
	Color        Field[string]
	Transmission *TransmissionType
	Status       *CarStatus
}

type CarFilter struct {
	BranchID           *string
	InstructorID       *string
	Status             *CarStatus
	AvailabilityStatus *AvailabilityStatus
	Limit              int
	Offset             int
}

type CarAssignment struct {
	InstructorID  string
	EffectiveFrom *time.Time
	Comment       *string
}

type CarAssignmentResult struct {
	CarID                string
	InstructorID         string
	PreviousInstructorID *string
	AssignedAt           time.Time
}
