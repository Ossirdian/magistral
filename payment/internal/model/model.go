package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Payment struct {
	ID        string
	SchoolID  string
	StudentID string
	Type      PaymentType
	Amount    decimal.Decimal
	Comment   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PaymentCreate struct {
	StudentID string
	Type      PaymentType
	Amount    decimal.Decimal
	Comment   *string
}

type PaymentPatch struct {
	Type       *PaymentType
	Amount     *decimal.Decimal
	Comment    *string
	CommentSet bool
}

type PaymentFilter struct {
	StudentID *string
	Type      *PaymentType
	Limit     int
	Offset    int
}

type StudentBalance struct {
	StudentID       string
	ContractTotal   decimal.Decimal
	PaidAmount      decimal.Decimal
	RemainingAmount decimal.Decimal
	UpdatedAt       time.Time
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}

type PaymentType string

const (
	PaymentTypeCash     PaymentType = "cash"
	PaymentTypeCard     PaymentType = "card"
	PaymentTypeTransfer PaymentType = "transfer"
	PaymentTypeOther    PaymentType = "other"
)
