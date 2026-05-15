package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/ossirdian/magistral/payment/internal/model"
	"github.com/ossirdian/magistral/payment/internal/repository"
	svc "github.com/ossirdian/magistral/payment/internal/service"
	"github.com/shopspring/decimal"
)

type studentValidatorStub struct {
	calls       int
	lastSchool  string
	lastStudent string
	err         error
}

func (s *studentValidatorStub) EnsureStudentBelongsToSchool(_ context.Context, schoolID, studentID string) error {
	s.calls++
	s.lastSchool = schoolID
	s.lastStudent = studentID
	return s.err
}

type paymentRepositoryStub struct {
	createFn         func(context.Context, string, model.PaymentCreate) (model.Payment, error)
	listFn           func(context.Context, string, model.PaymentFilter) ([]model.Payment, model.PageMeta, error)
	updateFn         func(context.Context, string, string, model.PaymentPatch) (model.Payment, error)
	deleteFn         func(context.Context, string, string) error
	getBalanceFn     func(context.Context, string, string) (model.StudentBalance, error)
	upsertContractFn func(context.Context, string, string, decimal.Decimal) (model.StudentBalance, error)
}

func (s *paymentRepositoryStub) GetByID(context.Context, string, string) (model.Payment, error) {
	panic("unexpected GetByID call")
}

func (s *paymentRepositoryStub) List(ctx context.Context, schoolID string, filter model.PaymentFilter) ([]model.Payment, model.PageMeta, error) {
	if s.listFn == nil {
		panic("unexpected List call")
	}
	return s.listFn(ctx, schoolID, filter)
}

func (s *paymentRepositoryStub) Create(ctx context.Context, schoolID string, input model.PaymentCreate) (model.Payment, error) {
	if s.createFn == nil {
		panic("unexpected Create call")
	}
	return s.createFn(ctx, schoolID, input)
}

func (s *paymentRepositoryStub) Update(ctx context.Context, schoolID, paymentID string, patch model.PaymentPatch) (model.Payment, error) {
	if s.updateFn == nil {
		panic("unexpected Update call")
	}
	return s.updateFn(ctx, schoolID, paymentID, patch)
}

func (s *paymentRepositoryStub) Delete(ctx context.Context, schoolID, paymentID string) error {
	if s.deleteFn == nil {
		panic("unexpected Delete call")
	}
	return s.deleteFn(ctx, schoolID, paymentID)
}

func (s *paymentRepositoryStub) GetBalance(ctx context.Context, schoolID, studentID string) (model.StudentBalance, error) {
	if s.getBalanceFn == nil {
		panic("unexpected GetBalance call")
	}
	return s.getBalanceFn(ctx, schoolID, studentID)
}

func (s *paymentRepositoryStub) UpsertContract(ctx context.Context, schoolID, studentID string, contractTotal decimal.Decimal) (model.StudentBalance, error) {
	if s.upsertContractFn == nil {
		panic("unexpected UpsertContract call")
	}
	return s.upsertContractFn(ctx, schoolID, studentID, contractTotal)
}

func TestCreateTrimsCommentAndValidatesStudent(t *testing.T) {
	t.Parallel()

	repo := &paymentRepositoryStub{
		createFn: func(_ context.Context, schoolID string, input model.PaymentCreate) (model.Payment, error) {
			if schoolID != "school-1" {
				t.Fatalf("unexpected schoolID: %s", schoolID)
			}
			if input.StudentID != "student-1" {
				t.Fatalf("unexpected studentID: %s", input.StudentID)
			}
			if input.Comment == nil || *input.Comment != "predoplata" {
				t.Fatalf("unexpected comment: %#v", input.Comment)
			}
			return model.Payment{StudentID: input.StudentID, Comment: input.Comment}, nil
		},
	}
	validator := &studentValidatorStub{}
	service := NewService(repo, validator)
	comment := "  predoplata  "

	payment, err := service.Create(context.Background(), "school-1", model.PaymentCreate{
		StudentID: "student-1",
		Type:      model.PaymentTypeCash,
		Amount:    decimal.RequireFromString("1000.50"),
		Comment:   &comment,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("expected validator call count 1, got %d", validator.calls)
	}
	if validator.lastSchool != "school-1" || validator.lastStudent != "student-1" {
		t.Fatalf("unexpected validator args: %+v", validator)
	}
	if payment.Comment == nil || *payment.Comment != "predoplata" {
		t.Fatalf("unexpected returned comment: %#v", payment.Comment)
	}
}

func TestCreateRejectsInvalidAmount(t *testing.T) {
	t.Parallel()

	service := NewService(&paymentRepositoryStub{}, &studentValidatorStub{})

	_, err := service.Create(context.Background(), "school-1", model.PaymentCreate{
		StudentID: "student-1",
		Type:      model.PaymentTypeCash,
		Amount:    decimal.Zero,
	})
	if !errors.Is(err, svc.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestListNormalizesStudentFilter(t *testing.T) {
	t.Parallel()

	repo := &paymentRepositoryStub{
		listFn: func(_ context.Context, schoolID string, filter model.PaymentFilter) ([]model.Payment, model.PageMeta, error) {
			if schoolID != "school-1" {
				t.Fatalf("unexpected schoolID: %s", schoolID)
			}
			if filter.StudentID == nil || *filter.StudentID != "student-1" {
				t.Fatalf("unexpected student filter: %#v", filter.StudentID)
			}
			if filter.Type == nil || *filter.Type != model.PaymentTypeCard {
				t.Fatalf("unexpected type filter: %#v", filter.Type)
			}
			return nil, model.PageMeta{Limit: 10, Offset: 5, Count: 1}, nil
		},
	}
	service := NewService(repo, &studentValidatorStub{})
	studentID := "  student-1  "
	paymentType := model.PaymentTypeCard

	_, _, err := service.List(context.Background(), "school-1", model.PaymentFilter{
		StudentID: &studentID,
		Type:      &paymentType,
		Limit:     10,
		Offset:    5,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
}

func TestUpdateTrimsComment(t *testing.T) {
	t.Parallel()

	repo := &paymentRepositoryStub{
		updateFn: func(_ context.Context, schoolID, paymentID string, patch model.PaymentPatch) (model.Payment, error) {
			if schoolID != "school-1" || paymentID != "payment-1" {
				t.Fatalf("unexpected identifiers: %s %s", schoolID, paymentID)
			}
			if !patch.CommentSet || patch.Comment == nil || *patch.Comment != "ispravleno" {
				t.Fatalf("unexpected patch comment: %#v", patch.Comment)
			}
			return model.Payment{ID: paymentID, Comment: patch.Comment}, nil
		},
	}
	service := NewService(repo, &studentValidatorStub{})
	comment := "  ispravleno  "

	payment, err := service.Update(context.Background(), "school-1", "payment-1", model.PaymentPatch{
		Comment:    &comment,
		CommentSet: true,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if payment.Comment == nil || *payment.Comment != "ispravleno" {
		t.Fatalf("unexpected returned comment: %#v", payment.Comment)
	}
}

func TestGetBalanceValidatesStudentBeforeRepository(t *testing.T) {
	t.Parallel()

	expected := model.StudentBalance{
		StudentID:       "student-1",
		ContractTotal:   decimal.RequireFromString("50000"),
		PaidAmount:      decimal.RequireFromString("20000"),
		RemainingAmount: decimal.RequireFromString("30000"),
	}
	repo := &paymentRepositoryStub{
		getBalanceFn: func(_ context.Context, schoolID, studentID string) (model.StudentBalance, error) {
			if schoolID != "school-1" || studentID != "student-1" {
				t.Fatalf("unexpected identifiers: %s %s", schoolID, studentID)
			}
			return expected, nil
		},
	}
	validator := &studentValidatorStub{}
	service := NewService(repo, validator)

	got, err := service.GetBalance(context.Background(), "school-1", "student-1")
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("expected validator call count 1, got %d", validator.calls)
	}
	if !got.RemainingAmount.Equal(expected.RemainingAmount) {
		t.Fatalf("unexpected balance: %+v", got)
	}
}

func TestUpsertContractRejectsNegativeTotal(t *testing.T) {
	t.Parallel()

	service := NewService(&paymentRepositoryStub{}, &studentValidatorStub{})

	_, err := service.UpsertContract(context.Background(), "school-1", "student-1", decimal.RequireFromString("-1"))
	if !errors.Is(err, svc.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUpsertContractPropagatesValidatorError(t *testing.T) {
	t.Parallel()

	expectedErr := repository.ErrNotFound
	validator := &studentValidatorStub{err: expectedErr}
	service := NewService(&paymentRepositoryStub{}, validator)

	_, err := service.UpsertContract(context.Background(), "school-1", "student-1", decimal.RequireFromString("100"))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
