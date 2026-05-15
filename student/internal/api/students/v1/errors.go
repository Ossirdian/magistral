package api

import (
	"context"
	"errors"

	studentsv1 "github.com/ossirdian/magistral/shared/pkg/students/v1"
	repo "github.com/ossirdian/magistral/student/internal/repository"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (a *api) NewError(ctx context.Context, err error) *studentsv1.DefaultErrorStatusCode {
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

	return &studentsv1.DefaultErrorStatusCode{
		StatusCode: statusCode,
		Response:   newErrorBody(ctx, code, message),
	}
}

func newErrorBody(_ context.Context, code, message string) studentsv1.ErrorBody {
	return studentsv1.ErrorBody{
		Error: studentsv1.ErrorObject{
			Code:    code,
			Message: message,
		},
		RequestID: "",
	}
}
