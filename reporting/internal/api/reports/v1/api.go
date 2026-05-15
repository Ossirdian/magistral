package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/reporting/internal/httpauth"
	"github.com/ossirdian/magistral/reporting/internal/service"
	reportingv1 "github.com/ossirdian/magistral/shared/pkg/reporting/v1"
)

var errNilReportingService = errors.New("reporting service is nil")

type api struct {
	reportingService service.ReportingService
}

func New(reportingService service.ReportingService) (*api, error) {
	if reportingService == nil {
		return nil, errNilReportingService
	}
	return &api{reportingService: reportingService}, nil
}

func NewServer(reportingService service.ReportingService, opts ...reportingv1.ServerOption) (*reportingv1.Server, error) {
	handler, err := New(reportingService)
	if err != nil {
		return nil, err
	}
	return reportingv1.NewServer(handler, opts...)
}

func (a *api) schoolID(ctx context.Context) (string, error) {
	schoolID, ok := httpauth.SchoolIDFromContext(ctx)
	if !ok || schoolID == "" {
		return "", service.ErrInvalidArgument
	}
	return schoolID, nil
}
