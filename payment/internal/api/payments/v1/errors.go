package api

import (
	"context"
	"errors"

	repo "github.com/ossirdian/magistral/payment/internal/repository"
	svc "github.com/ossirdian/magistral/payment/internal/service"
	paymentsv1 "github.com/ossirdian/magistral/shared/pkg/payments/v1"
)

func (a *api) NewError(ctx context.Context, err error) *paymentsv1.DefaultErrorStatusCode {
	statusCode := 500
	code := "internal_error"
	message := "internal server error"

	switch {
	case errors.Is(err, svc.ErrInvalidArgument):
		statusCode = 400
		code = "bad_request"
		message = "invalid request"
	case errors.Is(err, repo.ErrNotFound):
		statusCode = 404
		code = "not_found"
		message = "resource not found"
	case errors.Is(err, repo.ErrConflict):
		statusCode = 409
		code = "conflict"
		message = "resource conflict"
	}

	return &paymentsv1.DefaultErrorStatusCode{
		StatusCode: statusCode,
		Response:   newErrorBody(ctx, code, message),
	}
}

func newErrorBody(_ context.Context, code, message string) paymentsv1.ErrorBody {
	return paymentsv1.ErrorBody{
		Error:     paymentsv1.ErrorObject{Code: code, Message: message},
		RequestID: "",
	}
}
