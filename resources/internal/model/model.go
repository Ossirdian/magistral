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

type AvailabilityWindow struct {
	ID           string
	ResourceType ResourceType
	ResourceID   string
	Type         AvailabilityWindowType
	StartAt      time.Time
	EndAt        time.Time
	Comment      *string
}

type AvailabilityWindowInput struct {
	Type    AvailabilityWindowType
	StartAt time.Time
	EndAt   time.Time
	Comment *string
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

type ResourceType string

const (
	ResourceTypeInstructor ResourceType = "instructor"
	ResourceTypeCar        ResourceType = "car"
)

type AvailabilityStatus string

const (
	AvailabilityStatusAvailable   AvailabilityStatus = "available"
	AvailabilityStatusUnavailable AvailabilityStatus = "unavailable"
	AvailabilityStatusPartial     AvailabilityStatus = "partial"
)

type InstructorStatus string

const (
	InstructorStatusActive   InstructorStatus = "active"
	InstructorStatusInactive InstructorStatus = "inactive"
	InstructorStatusOnLeave  InstructorStatus = "on_leave"
)

type CarStatus string

const (
	CarStatusActive      CarStatus = "active"
	CarStatusInactive    CarStatus = "inactive"
	CarStatusMaintenance CarStatus = "maintenance"
)

type TransmissionType string

const (
	TransmissionTypeManual    TransmissionType = "manual"
	TransmissionTypeAutomatic TransmissionType = "automatic"
)

type AvailabilityWindowType string

const (
	AvailabilityWindowTypeWorkingHours AvailabilityWindowType = "working_hours"
	AvailabilityWindowTypeVacation     AvailabilityWindowType = "vacation"
	AvailabilityWindowTypeBlackout     AvailabilityWindowType = "blackout"
)
