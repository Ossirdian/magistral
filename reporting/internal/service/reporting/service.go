package reporting

import (
	"context"
	"sort"
	"strings"

	"github.com/ossirdian/magistral/reporting/internal/model"
	svc "github.com/ossirdian/magistral/reporting/internal/service"
	"github.com/shopspring/decimal"
)

type StudentReader interface {
	CountActiveStudents(ctx context.Context, schoolID string, filter model.ReportFilter) (int, error)
	StudentReport(ctx context.Context, schoolID string, filter model.ReportFilter) (model.StudentsReport, error)
	StudentIDs(ctx context.Context, schoolID string, filter model.ReportFilter) ([]string, error)
}

type LessonReader interface {
	CountLessons(ctx context.Context, schoolID string, filter model.ReportFilter) (int, error)
	LessonsReport(ctx context.Context, schoolID string, filter model.ReportFilter) (model.LessonsReport, error)
	InstructorLoads(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.EntityLoad, error)
	CarLoads(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.EntityLoad, error)
}

type ResourcesReader interface {
	CountInstructors(ctx context.Context, schoolID string, branchID *string) (int, error)
	CountCars(ctx context.Context, schoolID string, branchID *string) (int, error)
	ListInstructors(ctx context.Context, schoolID string, branchID *string) ([]model.InstructorRef, error)
	ListCars(ctx context.Context, schoolID string, branchID *string) ([]model.CarRef, error)
	InstructorAvailabilityMinutes(ctx context.Context, schoolID string, filter model.ReportFilter) (map[string]int, error)
	CarAvailabilityMinutes(ctx context.Context, schoolID string, filter model.ReportFilter) (map[string]int, error)
}

type PaymentReader interface {
	PaymentsTotal(ctx context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error)
	RemainingTotal(ctx context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error)
	PaymentTypeBreakdown(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.PaymentTypeAmount, error)
}

type Service struct {
	students  StudentReader
	lessons   LessonReader
	resources ResourcesReader
	payments  PaymentReader
}

func NewService(students StudentReader, lessons LessonReader, resources ResourcesReader, payments PaymentReader) *Service {
	return &Service{students: students, lessons: lessons, resources: resources, payments: payments}
}

func (s *Service) Dashboard(ctx context.Context, schoolID string, filter model.ReportFilter) (model.DashboardReport, error) {
	filter, err := normalizeFilter(schoolID, filter)
	if err != nil {
		return model.DashboardReport{}, err
	}

	filter, err = s.withStudentIDs(ctx, schoolID, filter)
	if err != nil {
		return model.DashboardReport{}, err
	}

	activeStudents, err := s.students.CountActiveStudents(ctx, schoolID, filter)
	if err != nil {
		return model.DashboardReport{}, err
	}
	lessonsTotal, err := s.lessons.CountLessons(ctx, schoolID, filter)
	if err != nil {
		return model.DashboardReport{}, err
	}
	instructorsTotal, err := s.resources.CountInstructors(ctx, schoolID, filter.BranchID)
	if err != nil {
		return model.DashboardReport{}, err
	}
	carsTotal, err := s.resources.CountCars(ctx, schoolID, filter.BranchID)
	if err != nil {
		return model.DashboardReport{}, err
	}
	paymentsTotal, err := s.payments.PaymentsTotal(ctx, schoolID, filter)
	if err != nil {
		return model.DashboardReport{}, err
	}
	remainingTotal, err := s.payments.RemainingTotal(ctx, schoolID, filter)
	if err != nil {
		return model.DashboardReport{}, err
	}

	return model.DashboardReport{
		ActiveStudents:          activeStudents,
		LessonsTotal:            lessonsTotal,
		InstructorsTotal:        instructorsTotal,
		CarsTotal:               carsTotal,
		PaymentsTotal:           paymentsTotal,
		ContractsRemainingTotal: remainingTotal,
	}, nil
}

func (s *Service) Students(ctx context.Context, schoolID string, filter model.ReportFilter) (model.StudentsReport, error) {
	filter, err := normalizeFilter(schoolID, filter)
	if err != nil {
		return model.StudentsReport{}, err
	}
	return s.students.StudentReport(ctx, schoolID, filter)
}

func (s *Service) Lessons(ctx context.Context, schoolID string, filter model.ReportFilter) (model.LessonsReport, error) {
	filter, err := normalizeFilter(schoolID, filter)
	if err != nil {
		return model.LessonsReport{}, err
	}
	filter, err = s.withStudentIDs(ctx, schoolID, filter)
	if err != nil {
		return model.LessonsReport{}, err
	}
	return s.lessons.LessonsReport(ctx, schoolID, filter)
}

func (s *Service) InstructorWorkload(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.InstructorWorkloadItem, error) {
	filter, err := normalizeFilter(schoolID, filter)
	if err != nil {
		return nil, err
	}
	filter, err = s.withStudentIDs(ctx, schoolID, filter)
	if err != nil {
		return nil, err
	}
	loads, err := s.lessons.InstructorLoads(ctx, schoolID, filter)
	if err != nil {
		return nil, err
	}
	refs, err := s.resources.ListInstructors(ctx, schoolID, filter.BranchID)
	if err != nil {
		return nil, err
	}
	availabilityByID, err := s.resources.InstructorAvailabilityMinutes(ctx, schoolID, filter)
	if err != nil {
		return nil, err
	}

	loadByID := make(map[string]model.EntityLoad, len(loads))
	for _, item := range loads {
		loadByID[item.ID] = item
	}

	result := make([]model.InstructorWorkloadItem, 0, max(len(refs), len(loads)))
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		load := loadByID[ref.ID]
		available := availabilityByID[ref.ID]
		result = append(result, model.InstructorWorkloadItem{
			InstructorID:     ref.ID,
			FullName:         ref.FullName,
			LessonsTotal:     load.LessonsTotal,
			TotalMinutes:     load.TotalMinutes,
			AvailableMinutes: available,
			IdleMinutes:      idleMinutes(available, load.TotalMinutes),
		})
		seen[ref.ID] = struct{}{}
	}
	for _, load := range loads {
		if _, ok := seen[load.ID]; ok {
			continue
		}
		available := availabilityByID[load.ID]
		result = append(result, model.InstructorWorkloadItem{
			InstructorID:     load.ID,
			FullName:         load.ID,
			LessonsTotal:     load.LessonsTotal,
			TotalMinutes:     load.TotalMinutes,
			AvailableMinutes: available,
			IdleMinutes:      idleMinutes(available, load.TotalMinutes),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalMinutes == result[j].TotalMinutes {
			return result[i].FullName < result[j].FullName
		}
		return result[i].TotalMinutes > result[j].TotalMinutes
	})
	return result, nil
}

func (s *Service) CarsUtilization(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.CarUtilizationItem, error) {
	filter, err := normalizeFilter(schoolID, filter)
	if err != nil {
		return nil, err
	}
	filter, err = s.withStudentIDs(ctx, schoolID, filter)
	if err != nil {
		return nil, err
	}
	loads, err := s.lessons.CarLoads(ctx, schoolID, filter)
	if err != nil {
		return nil, err
	}
	refs, err := s.resources.ListCars(ctx, schoolID, filter.BranchID)
	if err != nil {
		return nil, err
	}
	availabilityByID, err := s.resources.CarAvailabilityMinutes(ctx, schoolID, filter)
	if err != nil {
		return nil, err
	}

	loadByID := make(map[string]model.EntityLoad, len(loads))
	for _, item := range loads {
		loadByID[item.ID] = item
	}

	result := make([]model.CarUtilizationItem, 0, max(len(refs), len(loads)))
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		load := loadByID[ref.ID]
		available := availabilityByID[ref.ID]
		result = append(result, model.CarUtilizationItem{
			CarID:            ref.ID,
			PlateNumber:      ref.PlateNumber,
			DisplayName:      ref.DisplayName,
			LessonsTotal:     load.LessonsTotal,
			TotalMinutes:     load.TotalMinutes,
			AvailableMinutes: available,
			IdleMinutes:      idleMinutes(available, load.TotalMinutes),
		})
		seen[ref.ID] = struct{}{}
	}
	for _, load := range loads {
		if _, ok := seen[load.ID]; ok {
			continue
		}
		available := availabilityByID[load.ID]
		result = append(result, model.CarUtilizationItem{
			CarID:            load.ID,
			PlateNumber:      load.ID,
			DisplayName:      load.ID,
			LessonsTotal:     load.LessonsTotal,
			TotalMinutes:     load.TotalMinutes,
			AvailableMinutes: available,
			IdleMinutes:      idleMinutes(available, load.TotalMinutes),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalMinutes == result[j].TotalMinutes {
			return result[i].DisplayName < result[j].DisplayName
		}
		return result[i].TotalMinutes > result[j].TotalMinutes
	})
	return result, nil
}

func (s *Service) Finance(ctx context.Context, schoolID string, filter model.ReportFilter) (model.FinancialReport, error) {
	filter, err := normalizeFilter(schoolID, filter)
	if err != nil {
		return model.FinancialReport{}, err
	}

	filter, err = s.withStudentIDs(ctx, schoolID, filter)
	if err != nil {
		return model.FinancialReport{}, err
	}
	paymentsTotal, err := s.payments.PaymentsTotal(ctx, schoolID, filter)
	if err != nil {
		return model.FinancialReport{}, err
	}
	remainingTotal, err := s.payments.RemainingTotal(ctx, schoolID, filter)
	if err != nil {
		return model.FinancialReport{}, err
	}
	byType, err := s.payments.PaymentTypeBreakdown(ctx, schoolID, filter)
	if err != nil {
		return model.FinancialReport{}, err
	}

	return model.FinancialReport{
		PaymentsTotal:  paymentsTotal,
		RemainingTotal: remainingTotal,
		PaymentTypes:   byType,
	}, nil
}

func (s *Service) withStudentIDs(ctx context.Context, schoolID string, filter model.ReportFilter) (model.ReportFilter, error) {
	if filter.GroupID == nil && filter.BranchID == nil {
		return filter, nil
	}
	ids, err := s.students.StudentIDs(ctx, schoolID, filter)
	if err != nil {
		return model.ReportFilter{}, err
	}
	filter.StudentIDs = ids
	filter.StudentIDsFiltered = true
	return filter, nil
}

func normalizeFilter(schoolID string, filter model.ReportFilter) (model.ReportFilter, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.ReportFilter{}, svc.ErrInvalidArgument
	}
	if filter.BranchID != nil {
		value := strings.TrimSpace(*filter.BranchID)
		if value == "" {
			filter.BranchID = nil
		} else {
			filter.BranchID = &value
		}
	}
	if filter.GroupID != nil {
		value := strings.TrimSpace(*filter.GroupID)
		if value == "" {
			filter.GroupID = nil
		} else {
			filter.GroupID = &value
		}
	}
	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo) {
		return model.ReportFilter{}, svc.ErrInvalidArgument
	}
	return filter, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func idleMinutes(available, used int) int {
	if available <= used {
		return 0
	}
	return available - used
}

var _ svc.ReportingService = (*Service)(nil)
