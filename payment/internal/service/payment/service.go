package payment

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/payment/internal/model"
	"github.com/ossirdian/magistral/payment/internal/repository"
	svc "github.com/ossirdian/magistral/payment/internal/service"
	"github.com/shopspring/decimal"
)

type StudentValidator interface {
	EnsureStudentBelongsToSchool(ctx context.Context, schoolID, studentID string) error
}

type Service struct {
	repo             repository.PaymentRepository
	studentValidator StudentValidator
}

func NewService(repo repository.PaymentRepository, studentValidator StudentValidator) *Service {
	return &Service{repo: repo, studentValidator: studentValidator}
}

func (s *Service) Create(ctx context.Context, schoolID string, input model.PaymentCreate) (model.Payment, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.Payment{}, svc.ErrInvalidArgument
	}
	input.StudentID = strings.TrimSpace(input.StudentID)
	if input.StudentID == "" || !isValidPaymentType(input.Type) || !input.Amount.GreaterThan(decimal.Zero) {
		return model.Payment{}, svc.ErrInvalidArgument
	}
	if input.Comment != nil {
		comment := strings.TrimSpace(*input.Comment)
		input.Comment = &comment
	}
	if err := s.studentValidator.EnsureStudentBelongsToSchool(ctx, schoolID, input.StudentID); err != nil {
		return model.Payment{}, err
	}
	return s.repo.Create(ctx, schoolID, input)
}

func (s *Service) List(ctx context.Context, schoolID string, filter model.PaymentFilter) ([]model.Payment, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.StudentID != nil {
		value := strings.TrimSpace(*filter.StudentID)
		if value == "" {
			filter.StudentID = nil
		} else {
			filter.StudentID = &value
		}
	}
	if filter.Type != nil && !isValidPaymentType(*filter.Type) {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	return s.repo.List(ctx, schoolID, filter)
}

func (s *Service) Update(ctx context.Context, schoolID, paymentID string, patch model.PaymentPatch) (model.Payment, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(paymentID) == "" {
		return model.Payment{}, svc.ErrInvalidArgument
	}
	if patch.Type != nil && !isValidPaymentType(*patch.Type) {
		return model.Payment{}, svc.ErrInvalidArgument
	}
	if patch.Amount != nil && !patch.Amount.GreaterThan(decimal.Zero) {
		return model.Payment{}, svc.ErrInvalidArgument
	}
	if patch.CommentSet && patch.Comment != nil {
		comment := strings.TrimSpace(*patch.Comment)
		patch.Comment = &comment
	}
	return s.repo.Update(ctx, schoolID, paymentID, patch)
}

func (s *Service) Delete(ctx context.Context, schoolID, paymentID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(paymentID) == "" {
		return svc.ErrInvalidArgument
	}
	return s.repo.Delete(ctx, schoolID, paymentID)
}

func (s *Service) GetBalance(ctx context.Context, schoolID, studentID string) (model.StudentBalance, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(studentID) == "" {
		return model.StudentBalance{}, svc.ErrInvalidArgument
	}
	if err := s.studentValidator.EnsureStudentBelongsToSchool(ctx, schoolID, studentID); err != nil {
		return model.StudentBalance{}, err
	}
	return s.repo.GetBalance(ctx, schoolID, studentID)
}

func (s *Service) UpsertContract(ctx context.Context, schoolID, studentID string, contractTotal decimal.Decimal) (model.StudentBalance, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(studentID) == "" {
		return model.StudentBalance{}, svc.ErrInvalidArgument
	}
	if contractTotal.IsNegative() {
		return model.StudentBalance{}, svc.ErrInvalidArgument
	}
	if err := s.studentValidator.EnsureStudentBelongsToSchool(ctx, schoolID, studentID); err != nil {
		return model.StudentBalance{}, err
	}
	return s.repo.UpsertContract(ctx, schoolID, studentID, contractTotal)
}

func isValidPaymentType(value model.PaymentType) bool {
	switch value {
	case model.PaymentTypeCash, model.PaymentTypeCard, model.PaymentTypeTransfer, model.PaymentTypeOther:
		return true
	default:
		return false
	}
}
