package api

import (
	"context"
	"errors"

	repo "github.com/ossirdian/magistral/resources/internal/repository"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
)

func (a *api) NewError(ctx context.Context, err error) *resourcesv1.DefaultErrorStatusCode {
	statusCode := 500
	code := "internal_error"
	message := "internal server error"

	switch {
	case errors.Is(err, repo.ErrNotFound):
		statusCode = 404
		code = "not_found"
		message = "resource not found"
	case errors.Is(err, repo.ErrConflict):
		statusCode = 409
		code = "conflict"
		message = "resource conflict"
	}

	return &resourcesv1.DefaultErrorStatusCode{
		StatusCode: statusCode,
		Response:   newErrorBody(ctx, code, message),
	}
}

func newErrorBody(_ context.Context, code, message string) resourcesv1.ErrorBody {
	return resourcesv1.ErrorBody{
		Error: resourcesv1.ErrorObject{
			Code:    code,
			Message: message,
		},
		RequestID: "",
	}
}
