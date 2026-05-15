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
