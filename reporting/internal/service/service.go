package service

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/reporting/internal/model"
)

var ErrInvalidArgument = errors.New("invalid argument")

type ReportingService interface {
	Dashboard(ctx context.Context, schoolID string, filter model.ReportFilter) (model.DashboardReport, error)
	Students(ctx context.Context, schoolID string, filter model.ReportFilter) (model.StudentsReport, error)
	Lessons(ctx context.Context, schoolID string, filter model.ReportFilter) (model.LessonsReport, error)
	InstructorWorkload(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.InstructorWorkloadItem, error)
	CarsUtilization(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.CarUtilizationItem, error)
	Finance(ctx context.Context, schoolID string, filter model.ReportFilter) (model.FinancialReport, error)
}
