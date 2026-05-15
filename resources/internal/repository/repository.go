package repository

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

type CarRepository interface {
	GetByID(ctx context.Context, schoolID, carID string) (model.Car, error)
	List(ctx context.Context, schoolID string, filter model.CarFilter) ([]model.Car, model.PageMeta, error)
	Create(ctx context.Context, schoolID string, input model.CarCreate) (model.Car, error)
	Update(ctx context.Context, schoolID, carID string, patch model.CarPatch) (model.Car, error)
	Delete(ctx context.Context, schoolID, carID string) error
	GetAvailability(ctx context.Context, schoolID, carID string) ([]model.AvailabilityWindow, error)
	ReplaceAvailability(ctx context.Context, schoolID, carID string, windows []model.AvailabilityWindowInput) ([]model.AvailabilityWindow, error)
	AssignInstructor(ctx context.Context, schoolID, carID string, input model.CarAssignment) (model.CarAssignmentResult, error)
}

type InstructorRepository interface {
	GetByID(ctx context.Context, schoolID, instructorID string) (model.Instructor, error)
	List(ctx context.Context, schoolID string, filter model.InstructorFilter) ([]model.Instructor, model.PageMeta, error)
	Create(ctx context.Context, schoolID string, input model.InstructorCreate) (model.Instructor, error)
	Update(ctx context.Context, schoolID, instructorID string, patch model.InstructorPatch) (model.Instructor, error)
	Delete(ctx context.Context, schoolID, instructorID string) error
	GetAvailability(ctx context.Context, schoolID, instructorID string) ([]model.AvailabilityWindow, error)
	ReplaceAvailability(ctx context.Context, schoolID, instructorID string, windows []model.AvailabilityWindowInput) ([]model.AvailabilityWindow, error)
}
