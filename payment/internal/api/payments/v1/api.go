package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/payment/internal/httpauth"
	"github.com/ossirdian/magistral/payment/internal/service"
	paymentsv1 "github.com/ossirdian/magistral/shared/pkg/payments/v1"
)

var errNilPaymentService = errors.New("payment service is nil")

type api struct {
	paymentService service.PaymentService
}

func New(paymentService service.PaymentService) (*api, error) {
	if paymentService == nil {
		return nil, errNilPaymentService
	}
	return &api{paymentService: paymentService}, nil
}

func NewServer(paymentService service.PaymentService, opts ...paymentsv1.ServerOption) (*paymentsv1.Server, error) {
	handler, err := New(paymentService)
	if err != nil {
		return nil, err
	}
	return paymentsv1.NewServer(handler, opts...)
}

func (a *api) schoolID(ctx context.Context) (string, error) {
	schoolID, ok := httpauth.SchoolIDFromContext(ctx)
	if !ok || schoolID == "" {
		return "", service.ErrInvalidArgument
	}
	return schoolID, nil
}
