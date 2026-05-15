package converter

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/payment/internal/model"
	paymentsv1 "github.com/ossirdian/magistral/shared/pkg/payments/v1"
	"github.com/shopspring/decimal"
)

func ConvertPayment(item model.Payment) (paymentsv1.Payment, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return paymentsv1.Payment{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return paymentsv1.Payment{}, err
	}
	studentID, err := parseUUID(item.StudentID)
	if err != nil {
		return paymentsv1.Payment{}, err
	}
	return paymentsv1.Payment{
		ID:        id,
		SchoolID:  schoolID,
		StudentID: studentID,
		Type:      paymentsv1.PaymentType(item.Type),
		Amount:    item.Amount.InexactFloat64(),
		Comment:   ptrStringToOptNilString(item.Comment),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

func ConvertPayments(items []model.Payment) ([]paymentsv1.Payment, error) {
	result := make([]paymentsv1.Payment, 0, len(items))
	for _, item := range items {
		converted, err := ConvertPayment(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertBalance(item model.StudentBalance) (paymentsv1.StudentBalance, error) {
	studentID, err := parseUUID(item.StudentID)
	if err != nil {
		return paymentsv1.StudentBalance{}, err
	}
	return paymentsv1.StudentBalance{
		StudentID:       studentID,
		ContractTotal:   item.ContractTotal.InexactFloat64(),
		PaidAmount:      item.PaidAmount.InexactFloat64(),
		RemainingAmount: item.RemainingAmount.InexactFloat64(),
	}, nil
}

func ConvertPaymentCreateRequest(req *paymentsv1.PaymentCreateRequest) model.PaymentCreate {
	return model.PaymentCreate{
		StudentID: req.StudentID.String(),
		Type:      model.PaymentType(req.Type),
		Amount:    decimal.NewFromFloat(req.Amount),
		Comment:   optNilStringToPtr(req.Comment),
	}
}

func ConvertPaymentPatchRequest(req *paymentsv1.PaymentPatchRequest) model.PaymentPatch {
	comment, commentSet := optNilStringToPatch(req.Comment)
	return model.PaymentPatch{
		Type:       optPaymentTypeToPtr(req.Type),
		Amount:     optFloat64ToDecimalPtr(req.Amount),
		Comment:    comment,
		CommentSet: commentSet,
	}
}

func ConvertPaymentFilter(params paymentsv1.ListPaymentsParams) model.PaymentFilter {
	return model.PaymentFilter{
		StudentID: optUUIDToPtr(params.StudentID),
		Type:      optPaymentTypeToPtr(params.Type),
		Limit:     params.Limit.Or(0),
		Offset:    params.Offset.Or(0),
	}
}

func ConvertMeta(meta model.PageMeta) paymentsv1.Meta {
	return paymentsv1.Meta{Limit: meta.Limit, Offset: meta.Offset, Count: meta.Count}
}

func optUUIDToPtr(v paymentsv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optPaymentTypeToPtr(v paymentsv1.OptPaymentType) *model.PaymentType {
	if value, ok := v.Get(); ok {
		converted := model.PaymentType(value)
		return &converted
	}
	return nil
}

func optFloat64ToDecimalPtr(v paymentsv1.OptFloat64) *decimal.Decimal {
	if value, ok := v.Get(); ok {
		converted := decimal.NewFromFloat(value)
		return &converted
	}
	return nil
}

func optNilStringToPtr(v paymentsv1.OptNilString) *string {
	if !v.IsSet() || v.IsNull() {
		return nil
	}
	value, ok := v.Get()
	if !ok {
		return nil
	}
	return &value
}

func optNilStringToPatch(v paymentsv1.OptNilString) (*string, bool) {
	if !v.IsSet() {
		return nil, false
	}
	if v.IsNull() {
		return nil, true
	}
	value, ok := v.Get()
	if !ok {
		return nil, true
	}
	return &value, true
}

func ptrStringToOptNilString(v *string) paymentsv1.OptNilString {
	if v == nil {
		return paymentsv1.OptNilString{}
	}
	var result paymentsv1.OptNilString
	result.SetTo(*v)
	return result
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}
