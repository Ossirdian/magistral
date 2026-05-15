package service

import (
	"context"

	"github.com/ossirdian/magistral/payment/internal/model"
	"github.com/shopspring/decimal"
)

type PaymentService interface {
	Create(ctx context.Context, schoolID string, input model.PaymentCreate) (model.Payment, error)
	List(ctx context.Context, schoolID string, filter model.PaymentFilter) ([]model.Payment, model.PageMeta, error)
	Update(ctx context.Context, schoolID, paymentID string, patch model.PaymentPatch) (model.Payment, error)
	Delete(ctx context.Context, schoolID, paymentID string) error
	GetBalance(ctx context.Context, schoolID, studentID string) (model.StudentBalance, error)
	UpsertContract(ctx context.Context, schoolID, studentID string, contractTotal decimal.Decimal) (model.StudentBalance, error)
}
