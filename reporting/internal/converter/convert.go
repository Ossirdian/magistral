package converter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/reporting/internal/model"
	reportingv1 "github.com/ossirdian/magistral/shared/pkg/reporting/v1"
)

func ConvertFilter(branchID, groupID reportingv1.OptUUID, month reportingv1.OptString, dateFrom, dateTo reportingv1.OptDateTime) model.ReportFilter {
	from := optDateTimeToPtr(dateFrom)
	to := optDateTimeToPtr(dateTo)
	if from == nil && to == nil {
		if value, ok := month.Get(); ok && value != "" {
			if parsed, err := time.Parse("2006-01", value); err == nil {
				start := parsed.UTC()
				end := start.AddDate(0, 1, 0)
				from = &start
				to = &end
			}
		}
	}
	return model.ReportFilter{
		BranchID: optUUIDToPtr(branchID),
		GroupID:  optUUIDToPtr(groupID),
		DateFrom: from,
		DateTo:   to,
	}
}

func ConvertDashboardReport(item model.DashboardReport) reportingv1.DashboardReport {
	return reportingv1.DashboardReport{
		ActiveStudents:          int64(item.ActiveStudents),
		LessonsTotal:            int64(item.LessonsTotal),
		InstructorsTotal:        int64(item.InstructorsTotal),
		CarsTotal:               int64(item.CarsTotal),
		PaymentsTotal:           item.PaymentsTotal.InexactFloat64(),
		ContractsRemainingTotal: item.ContractsRemainingTotal.InexactFloat64(),
	}
}

func ConvertStudentsReport(item model.StudentsReport) reportingv1.StudentsReport {
	statuses := make([]reportingv1.StatusCount, 0, len(item.Statuses))
	for _, status := range item.Statuses {
		statuses = append(statuses, reportingv1.StatusCount{Status: status.Status, Count: int64(status.Count)})
	}
	return reportingv1.StudentsReport{
		Total:    int64(item.Total),
		Active:   int64(item.Active),
		New:      int64(item.New),
		Statuses: statuses,
	}
}

func ConvertLessonsReport(item model.LessonsReport) reportingv1.LessonsReport {
	return reportingv1.LessonsReport{
		Total:     int64(item.Total),
		Scheduled: int64(item.Scheduled),
		Completed: int64(item.Completed),
		Cancelled: int64(item.Cancelled),
		Missed:    int64(item.Missed),
	}
}

func ConvertInstructorWorkload(items []model.InstructorWorkloadItem) ([]reportingv1.InstructorWorkloadItem, error) {
	result := make([]reportingv1.InstructorWorkloadItem, 0, len(items))
	for _, item := range items {
		id, err := parseUUID(item.InstructorID)
		if err != nil {
			return nil, err
		}
		result = append(result, reportingv1.InstructorWorkloadItem{
			InstructorID:     id,
			FullName:         item.FullName,
			LessonsTotal:     int64(item.LessonsTotal),
			TotalMinutes:     int64(item.TotalMinutes),
			AvailableMinutes: int64(item.AvailableMinutes),
			IdleMinutes:      int64(item.IdleMinutes),
		})
	}
	return result, nil
}

func ConvertCarsUtilization(items []model.CarUtilizationItem) ([]reportingv1.CarUtilizationItem, error) {
	result := make([]reportingv1.CarUtilizationItem, 0, len(items))
	for _, item := range items {
		id, err := parseUUID(item.CarID)
		if err != nil {
			return nil, err
		}
		result = append(result, reportingv1.CarUtilizationItem{
			CarID:            id,
			PlateNumber:      item.PlateNumber,
			DisplayName:      item.DisplayName,
			LessonsTotal:     int64(item.LessonsTotal),
			TotalMinutes:     int64(item.TotalMinutes),
			AvailableMinutes: int64(item.AvailableMinutes),
			IdleMinutes:      int64(item.IdleMinutes),
		})
	}
	return result, nil
}

func ConvertFinancialReport(item model.FinancialReport) reportingv1.FinancialReport {
	types := make([]reportingv1.PaymentTypeAmount, 0, len(item.PaymentTypes))
	for _, row := range item.PaymentTypes {
		types = append(types, reportingv1.PaymentTypeAmount{Type: row.Type, Amount: row.Amount.InexactFloat64()})
	}
	return reportingv1.FinancialReport{
		PaymentsTotal:  item.PaymentsTotal.InexactFloat64(),
		RemainingTotal: item.RemainingTotal.InexactFloat64(),
		PaymentTypes:   types,
	}
}

func optUUIDToPtr(v reportingv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optDateTimeToPtr(v reportingv1.OptDateTime) *time.Time {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}
