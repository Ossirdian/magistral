package converter

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/payment/internal/model"
	paymentsv1 "github.com/ossirdian/magistral/shared/pkg/payments/v1"
	"github.com/shopspring/decimal"
)

func TestConvertPaymentCreateRequest(t *testing.T) {
	t.Parallel()

	studentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	req := &paymentsv1.PaymentCreateRequest{
		StudentID: studentID,
		Type:      paymentsv1.PaymentTypeTransfer,
		Amount:    1250.75,
		Comment:   paymentsv1.NewOptNilString("predoplata"),
	}

	got := ConvertPaymentCreateRequest(req)

	if got.StudentID != studentID.String() {
		t.Fatalf("unexpected studentID: %s", got.StudentID)
	}
	if got.Type != model.PaymentTypeTransfer {
		t.Fatalf("unexpected type: %s", got.Type)
	}
	if !got.Amount.Equal(decimal.RequireFromString("1250.75")) {
		t.Fatalf("unexpected amount: %s", got.Amount)
	}
	if got.Comment == nil || *got.Comment != "predoplata" {
		t.Fatalf("unexpected comment: %#v", got.Comment)
	}
}

func TestConvertPaymentPatchRequestPreservesNullComment(t *testing.T) {
	t.Parallel()

	var comment paymentsv1.OptNilString
	comment.SetToNull()
	req := &paymentsv1.PaymentPatchRequest{
		Type:    paymentsv1.NewOptPaymentType(paymentsv1.PaymentTypeCard),
		Amount:  paymentsv1.NewOptFloat64(1500.25),
		Comment: comment,
	}

	got := ConvertPaymentPatchRequest(req)

	if got.Type == nil || *got.Type != model.PaymentTypeCard {
		t.Fatalf("unexpected type: %#v", got.Type)
	}
	if got.Amount == nil || !got.Amount.Equal(decimal.RequireFromString("1500.25")) {
		t.Fatalf("unexpected amount: %#v", got.Amount)
	}
	if !got.CommentSet {
		t.Fatal("expected CommentSet to be true")
	}
	if got.Comment != nil {
		t.Fatalf("expected nil comment, got %#v", got.Comment)
	}
}

func TestConvertPaymentMapsDecimalAndComment(t *testing.T) {
	t.Parallel()

	comment := "ispravleno"
	now := time.Unix(1710000000, 0).UTC()
	item := model.Payment{
		ID:        "22222222-2222-2222-2222-222222222222",
		SchoolID:  "33333333-3333-3333-3333-333333333333",
		StudentID: "44444444-4444-4444-4444-444444444444",
		Type:      model.PaymentTypeCash,
		Amount:    decimal.RequireFromString("999.99"),
		Comment:   &comment,
		CreatedAt: now,
		UpdatedAt: now,
	}

	got, err := ConvertPayment(item)
	if err != nil {
		t.Fatalf("ConvertPayment returned error: %v", err)
	}
	if got.ID.String() != item.ID || got.SchoolID.String() != item.SchoolID || got.StudentID.String() != item.StudentID {
		t.Fatalf("unexpected uuid mapping: %+v", got)
	}
	if got.Amount != 999.99 {
		t.Fatalf("unexpected amount: %v", got.Amount)
	}
	value, ok := got.Comment.Get()
	if !ok || value != comment {
		t.Fatalf("unexpected comment: %#v", got.Comment)
	}
}

func TestConvertPaymentFailsOnInvalidUUID(t *testing.T) {
	t.Parallel()

	_, err := ConvertPayment(model.Payment{ID: "bad-uuid"})
	if err == nil {
		t.Fatal("expected error for invalid uuid")
	}
}

func TestConvertPaymentFilter(t *testing.T) {
	t.Parallel()

	studentID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	params := paymentsv1.ListPaymentsParams{
		StudentID: paymentsv1.NewOptUUID(studentID),
		Type:      paymentsv1.NewOptPaymentType(paymentsv1.PaymentTypeOther),
		Limit:     paymentsv1.NewOptInt(20),
		Offset:    paymentsv1.NewOptInt(40),
	}

	got := ConvertPaymentFilter(params)

	if got.StudentID == nil || *got.StudentID != studentID.String() {
		t.Fatalf("unexpected student filter: %#v", got.StudentID)
	}
	if got.Type == nil || *got.Type != model.PaymentTypeOther {
		t.Fatalf("unexpected type filter: %#v", got.Type)
	}
	if got.Limit != 20 || got.Offset != 40 {
		t.Fatalf("unexpected pagination: %+v", got)
	}
}
