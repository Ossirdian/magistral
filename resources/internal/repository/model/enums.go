package model

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
