package model

import "time"

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
