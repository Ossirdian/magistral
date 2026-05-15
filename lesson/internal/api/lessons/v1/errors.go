package api

import (
	"context"
	"errors"

	repo "github.com/ossirdian/magistral/lesson/internal/repository"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
	lessonsv1 "github.com/ossirdian/magistral/shared/pkg/lessons/v1"
)

func (a *api) NewError(ctx context.Context, err error) *lessonsv1.DefaultErrorStatusCode {
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

	return &lessonsv1.DefaultErrorStatusCode{
		StatusCode: statusCode,
		Response:   newErrorBody(ctx, code, message),
	}
}

func newErrorBody(_ context.Context, code, message string) lessonsv1.ErrorBody {
	return lessonsv1.ErrorBody{
		Error: lessonsv1.ErrorObject{
			Code:    code,
			Message: message,
		},
		RequestID: "",
	}
}
