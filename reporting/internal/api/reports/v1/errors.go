package api

import (
	"context"
	"errors"

	svc "github.com/ossirdian/magistral/reporting/internal/service"
	reportingv1 "github.com/ossirdian/magistral/shared/pkg/reporting/v1"
)

func (a *api) NewError(ctx context.Context, err error) *reportingv1.DefaultErrorStatusCode {
	statusCode := 500
	code := "internal_error"
	message := "internal server error"

	switch {
	case errors.Is(err, svc.ErrInvalidArgument):
		statusCode = 400
		code = "bad_request"
		message = "invalid request"
	}

	return &reportingv1.DefaultErrorStatusCode{
		StatusCode: statusCode,
		Response:   newErrorBody(ctx, code, message),
	}
}

func newErrorBody(_ context.Context, code, message string) reportingv1.ErrorBody {
	return reportingv1.ErrorBody{
		Error:     reportingv1.ErrorObject{Code: code, Message: message},
		RequestID: "",
	}
}
