package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ossirdian/magistral/reporting/internal/model"
	svc "github.com/ossirdian/magistral/reporting/internal/service"
	"github.com/shopspring/decimal"
)

type studentReaderStub struct {
	studentIDsFn          func(context.Context, string, model.ReportFilter) ([]string, error)
	countActiveStudentsFn func(context.Context, string, model.ReportFilter) (int, error)
	studentReportFn       func(context.Context, string, model.ReportFilter) (model.StudentsReport, error)
}

type lessonReaderStub struct {
	countLessonsFn    func(context.Context, string, model.ReportFilter) (int, error)
	lessonsReportFn   func(context.Context, string, model.ReportFilter) (model.LessonsReport, error)
	instructorLoadsFn func(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error)
	carLoadsFn        func(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error)
}

func (s *studentReaderStub) StudentIDs(ctx context.Context, schoolID string, filter model.ReportFilter) ([]string, error) {
	return s.studentIDsFn(ctx, schoolID, filter)
}
func (s *studentReaderStub) CountActiveStudents(ctx context.Context, schoolID string, filter model.ReportFilter) (int, error) {
	return s.countActiveStudentsFn(ctx, schoolID, filter)
}
func (s *studentReaderStub) StudentReport(ctx context.Context, schoolID string, filter model.ReportFilter) (model.StudentsReport, error) {
	return s.studentReportFn(ctx, schoolID, filter)
}
func (s *lessonReaderStub) CountLessons(ctx context.Context, schoolID string, filter model.ReportFilter) (int, error) {
	return s.countLessonsFn(ctx, schoolID, filter)
}
func (s *lessonReaderStub) LessonsReport(ctx context.Context, schoolID string, filter model.ReportFilter) (model.LessonsReport, error) {
	return s.lessonsReportFn(ctx, schoolID, filter)
}
func (s *lessonReaderStub) InstructorLoads(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.EntityLoad, error) {
	return s.instructorLoadsFn(ctx, schoolID, filter)
}
func (s *lessonReaderStub) CarLoads(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.EntityLoad, error) {
	return s.carLoadsFn(ctx, schoolID, filter)
}

type resourcesReaderStub struct {
	countInstructorsFn func(context.Context, string, *string) (int, error)
	countCarsFn        func(context.Context, string, *string) (int, error)
	listInstructorsFn  func(context.Context, string, *string) ([]model.InstructorRef, error)
	listCarsFn         func(context.Context, string, *string) ([]model.CarRef, error)
}

func (s *resourcesReaderStub) CountInstructors(ctx context.Context, schoolID string, branchID *string) (int, error) {
	return s.countInstructorsFn(ctx, schoolID, branchID)
}
func (s *resourcesReaderStub) CountCars(ctx context.Context, schoolID string, branchID *string) (int, error) {
	return s.countCarsFn(ctx, schoolID, branchID)
}
func (s *resourcesReaderStub) ListInstructors(ctx context.Context, schoolID string, branchID *string) ([]model.InstructorRef, error) {
	return s.listInstructorsFn(ctx, schoolID, branchID)
}
func (s *resourcesReaderStub) ListCars(ctx context.Context, schoolID string, branchID *string) ([]model.CarRef, error) {
	return s.listCarsFn(ctx, schoolID, branchID)
}

type paymentReaderStub struct {
	paymentsTotalFn    func(context.Context, string, model.ReportFilter) (decimal.Decimal, error)
	remainingTotalFn   func(context.Context, string, model.ReportFilter) (decimal.Decimal, error)
	paymentBreakdownFn func(context.Context, string, model.ReportFilter) ([]model.PaymentTypeAmount, error)
}

func (s *paymentReaderStub) PaymentsTotal(ctx context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
	return s.paymentsTotalFn(ctx, schoolID, filter)
}
func (s *paymentReaderStub) RemainingTotal(ctx context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
	return s.remainingTotalFn(ctx, schoolID, filter)
}
func (s *paymentReaderStub) PaymentTypeBreakdown(ctx context.Context, schoolID string, filter model.ReportFilter) ([]model.PaymentTypeAmount, error) {
	return s.paymentBreakdownFn(ctx, schoolID, filter)
}

func TestDashboardUsesBranchStudentIDsForFinance(t *testing.T) {
	t.Parallel()

	branchID := "branch-1"
	service := NewService(
		&studentReaderStub{
			studentIDsFn: func(_ context.Context, schoolID string, filter model.ReportFilter) ([]string, error) {
				if schoolID != "school-1" || filter.BranchID == nil || *filter.BranchID != branchID {
					t.Fatalf("unexpected branch lookup args: %s %+v", schoolID, filter.BranchID)
				}
				return []string{"student-1", "student-2"}, nil
			},
			countActiveStudentsFn: func(context.Context, string, model.ReportFilter) (int, error) {
				return 5, nil
			},
			studentReportFn: func(context.Context, string, model.ReportFilter) (model.StudentsReport, error) {
				return model.StudentsReport{}, nil
			},
		},
		&lessonReaderStub{
			countLessonsFn: func(context.Context, string, model.ReportFilter) (int, error) { return 11, nil },
			lessonsReportFn: func(context.Context, string, model.ReportFilter) (model.LessonsReport, error) {
				return model.LessonsReport{}, nil
			},
			instructorLoadsFn: func(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error) { return nil, nil },
			carLoadsFn:        func(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error) { return nil, nil },
		},
		&resourcesReaderStub{
			countInstructorsFn: func(context.Context, string, *string) (int, error) { return 3, nil },
			countCarsFn:        func(context.Context, string, *string) (int, error) { return 2, nil },
			listInstructorsFn:  func(context.Context, string, *string) ([]model.InstructorRef, error) { return nil, nil },
			listCarsFn:         func(context.Context, string, *string) ([]model.CarRef, error) { return nil, nil },
		},
		&paymentReaderStub{
			paymentsTotalFn: func(_ context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
				if schoolID != "school-1" || len(filter.StudentIDs) != 2 {
					t.Fatalf("unexpected payment args: %s %+v", schoolID, filter.StudentIDs)
				}
				return decimal.RequireFromString("1000"), nil
			},
			remainingTotalFn: func(_ context.Context, schoolID string, filter model.ReportFilter) (decimal.Decimal, error) {
				if schoolID != "school-1" || len(filter.StudentIDs) != 2 {
					t.Fatalf("unexpected remaining args: %s %+v", schoolID, filter.StudentIDs)
				}
				return decimal.RequireFromString("250"), nil
			},
			paymentBreakdownFn: func(context.Context, string, model.ReportFilter) ([]model.PaymentTypeAmount, error) {
				return nil, nil
			},
		},
	)

	report, err := service.Dashboard(context.Background(), "school-1", model.ReportFilter{BranchID: &branchID})
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}
	if !report.PaymentsTotal.Equal(decimal.RequireFromString("1000")) || !report.ContractsRemainingTotal.Equal(decimal.RequireFromString("250")) {
		t.Fatalf("unexpected dashboard totals: %+v", report)
	}
}

func TestFinanceRejectsInvalidDateRange(t *testing.T) {
	t.Parallel()

	service := NewService(&studentReaderStub{}, &lessonReaderStub{}, &resourcesReaderStub{}, &paymentReaderStub{})
	dateFrom := time.Now().UTC()
	dateTo := dateFrom.Add(-time.Hour)

	_, err := service.Finance(context.Background(), "school-1", model.ReportFilter{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	})
	if !errors.Is(err, svc.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestInstructorWorkloadMergesRefsAndLoads(t *testing.T) {
	t.Parallel()

	service := NewService(
		&studentReaderStub{
			studentIDsFn: func(context.Context, string, model.ReportFilter) ([]string, error) { return nil, nil },
			countActiveStudentsFn: func(context.Context, string, model.ReportFilter) (int, error) {
				return 0, nil
			},
			studentReportFn: func(context.Context, string, model.ReportFilter) (model.StudentsReport, error) {
				return model.StudentsReport{}, nil
			},
		},
		&lessonReaderStub{
			countLessonsFn: func(context.Context, string, model.ReportFilter) (int, error) { return 0, nil },
			lessonsReportFn: func(context.Context, string, model.ReportFilter) (model.LessonsReport, error) {
				return model.LessonsReport{}, nil
			},
			instructorLoadsFn: func(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error) {
				return []model.EntityLoad{{ID: "inst-1", LessonsTotal: 4, TotalMinutes: 360}}, nil
			},
			carLoadsFn: func(context.Context, string, model.ReportFilter) ([]model.EntityLoad, error) { return nil, nil },
		},
		&resourcesReaderStub{
			countInstructorsFn: func(context.Context, string, *string) (int, error) { return 0, nil },
			countCarsFn:        func(context.Context, string, *string) (int, error) { return 0, nil },
			listInstructorsFn: func(context.Context, string, *string) ([]model.InstructorRef, error) {
				return []model.InstructorRef{{ID: "inst-1", FullName: "Ivan Instructor"}, {ID: "inst-2", FullName: "Zero Load"}}, nil
			},
			listCarsFn: func(context.Context, string, *string) ([]model.CarRef, error) { return nil, nil },
		},
		&paymentReaderStub{
			paymentsTotalFn: func(context.Context, string, model.ReportFilter) (decimal.Decimal, error) {
				return decimal.Zero, nil
			},
			remainingTotalFn: func(context.Context, string, model.ReportFilter) (decimal.Decimal, error) {
				return decimal.Zero, nil
			},
			paymentBreakdownFn: func(context.Context, string, model.ReportFilter) ([]model.PaymentTypeAmount, error) {
				return nil, nil
			},
		},
	)

	items, err := service.InstructorWorkload(context.Background(), "school-1", model.ReportFilter{})
	if err != nil {
		t.Fatalf("InstructorWorkload returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].InstructorID != "inst-1" || items[0].TotalMinutes != 360 {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].InstructorID != "inst-2" || items[1].LessonsTotal != 0 {
		t.Fatalf("unexpected zero-load item: %+v", items[1])
	}
}
