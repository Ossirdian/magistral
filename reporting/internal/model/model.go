package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ReportFilter struct {
	BranchID           *string
	GroupID            *string
	DateFrom           *time.Time
	DateTo             *time.Time
	StudentIDs         []string
	StudentIDsFiltered bool
}

type StatusCount struct {
	Status string
	Count  int
}

type PaymentTypeAmount struct {
	Type   string
	Amount decimal.Decimal
}

type DashboardReport struct {
	ActiveStudents          int
	LessonsTotal            int
	InstructorsTotal        int
	CarsTotal               int
	PaymentsTotal           decimal.Decimal
	ContractsRemainingTotal decimal.Decimal
}

type StudentsReport struct {
	Total    int
	Active   int
	New      int
	Statuses []StatusCount
}

type LessonsReport struct {
	Total     int
	Scheduled int
	Completed int
	Cancelled int
	Missed    int
}

type InstructorWorkloadItem struct {
	InstructorID     string
	FullName         string
	LessonsTotal     int
	TotalMinutes     int
	AvailableMinutes int
	IdleMinutes      int
}

type CarUtilizationItem struct {
	CarID            string
	PlateNumber      string
	DisplayName      string
	LessonsTotal     int
	TotalMinutes     int
	AvailableMinutes int
	IdleMinutes      int
}

type FinancialReport struct {
	PaymentsTotal  decimal.Decimal
	RemainingTotal decimal.Decimal
	PaymentTypes   []PaymentTypeAmount
}

type EntityLoad struct {
	ID           string
	LessonsTotal int
	TotalMinutes int
}

type InstructorRef struct {
	ID       string
	FullName string
}

type CarRef struct {
	ID          string
	PlateNumber string
	DisplayName string
}
