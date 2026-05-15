package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/resources/internal/httpauth"
	"github.com/ossirdian/magistral/resources/internal/service"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
)

var (
	errNilCarService        = errors.New("car service is nil")
	errNilInstructorService = errors.New("instructor service is nil")
	errEmptySchoolID        = errors.New("school id is empty")
)

type api struct {
	defaultSchoolID   string
	carService        service.CarService
	instructorService service.InstructorService
}

func New(carService service.CarService, instructorService service.InstructorService, schoolID string) (*api, error) {
	switch {
	case carService == nil:
		return nil, errNilCarService
	case instructorService == nil:
		return nil, errNilInstructorService
	case schoolID == "":
		return nil, errEmptySchoolID
	}

	return &api{
		defaultSchoolID:   schoolID,
		carService:        carService,
		instructorService: instructorService,
	}, nil
}

func NewServer(carService service.CarService, instructorService service.InstructorService, schoolID string, opts ...resourcesv1.ServerOption) (*resourcesv1.Server, error) {
	handler, err := New(carService, instructorService, schoolID)
	if err != nil {
		return nil, err
	}

	return resourcesv1.NewServer(handler, opts...)
}

func (a *api) schoolID(ctx context.Context) string {
	if schoolID, ok := httpauth.SchoolIDFromContext(ctx); ok {
		return schoolID
	}

	return a.defaultSchoolID
}
