package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/reporting/internal/converter"
	svc "github.com/ossirdian/magistral/reporting/internal/service"
	reportingv1 "github.com/ossirdian/magistral/shared/pkg/reporting/v1"
)

func (a *api) GetDashboardReport(ctx context.Context, params reportingv1.GetDashboardReportParams) (reportingv1.GetDashboardReportRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := reportingv1.GetDashboardReportUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.reportingService.Dashboard(ctx, schoolID, converter.ConvertFilter(params.BranchID, params.GroupID, params.Month, params.DateFrom, params.DateTo))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := reportingv1.GetDashboardReportBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}
	return &reportingv1.DashboardReportResponse{Data: converter.ConvertDashboardReport(item)}, nil
}

func (a *api) GetStudentsReport(ctx context.Context, params reportingv1.GetStudentsReportParams) (reportingv1.GetStudentsReportRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := reportingv1.GetStudentsReportUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.reportingService.Students(ctx, schoolID, converter.ConvertFilter(params.BranchID, params.GroupID, params.Month, reportingv1.OptDateTime{}, reportingv1.OptDateTime{}))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := reportingv1.GetStudentsReportBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}
	return &reportingv1.StudentsReportResponse{Data: converter.ConvertStudentsReport(item)}, nil
}

func (a *api) GetLessonsReport(ctx context.Context, params reportingv1.GetLessonsReportParams) (reportingv1.GetLessonsReportRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := reportingv1.GetLessonsReportUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.reportingService.Lessons(ctx, schoolID, converter.ConvertFilter(params.BranchID, params.GroupID, params.Month, params.DateFrom, params.DateTo))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := reportingv1.GetLessonsReportBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}
	return &reportingv1.LessonsReportResponse{Data: converter.ConvertLessonsReport(item)}, nil
}

func (a *api) GetInstructorWorkloadReport(ctx context.Context, params reportingv1.GetInstructorWorkloadReportParams) (reportingv1.GetInstructorWorkloadReportRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := reportingv1.GetInstructorWorkloadReportUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, err := a.reportingService.InstructorWorkload(ctx, schoolID, converter.ConvertFilter(params.BranchID, params.GroupID, params.Month, params.DateFrom, params.DateTo))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := reportingv1.GetInstructorWorkloadReportBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}
	data, err := converter.ConvertInstructorWorkload(items)
	if err != nil {
		return nil, err
	}
	return &reportingv1.InstructorWorkloadResponse{Data: data}, nil
}

func (a *api) GetCarsUtilizationReport(ctx context.Context, params reportingv1.GetCarsUtilizationReportParams) (reportingv1.GetCarsUtilizationReportRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := reportingv1.GetCarsUtilizationReportUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, err := a.reportingService.CarsUtilization(ctx, schoolID, converter.ConvertFilter(params.BranchID, params.GroupID, params.Month, params.DateFrom, params.DateTo))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := reportingv1.GetCarsUtilizationReportBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}
	data, err := converter.ConvertCarsUtilization(items)
	if err != nil {
		return nil, err
	}
	return &reportingv1.CarsUtilizationResponse{Data: data}, nil
}

func (a *api) GetFinanceReport(ctx context.Context, params reportingv1.GetFinanceReportParams) (reportingv1.GetFinanceReportRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := reportingv1.GetFinanceReportUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.reportingService.Finance(ctx, schoolID, converter.ConvertFilter(params.BranchID, params.GroupID, params.Month, params.DateFrom, params.DateTo))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := reportingv1.GetFinanceReportBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}
	return &reportingv1.FinancialReportResponse{Data: converter.ConvertFinancialReport(item)}, nil
}
