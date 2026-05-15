package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/payment/internal/converter"
	repo "github.com/ossirdian/magistral/payment/internal/repository"
	svc "github.com/ossirdian/magistral/payment/internal/service"
	paymentsv1 "github.com/ossirdian/magistral/shared/pkg/payments/v1"
	"github.com/shopspring/decimal"
)

func (a *api) CreatePayment(ctx context.Context, req *paymentsv1.PaymentCreateRequest) (paymentsv1.CreatePaymentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := paymentsv1.CreatePaymentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
		return &res, nil
	}

	item, err := a.paymentService.Create(ctx, schoolID, converter.ConvertPaymentCreateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := paymentsv1.CreatePaymentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound), errors.Is(err, repo.ErrConflict):
			res := paymentsv1.CreatePaymentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertPayment(item)
	if err != nil {
		return nil, err
	}
	return &paymentsv1.PaymentResponse{Data: data}, nil
}

func (a *api) ListPayments(ctx context.Context, params paymentsv1.ListPaymentsParams) (paymentsv1.ListPaymentsRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := paymentsv1.ListPaymentsUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.paymentService.List(ctx, schoolID, converter.ConvertPaymentFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := paymentsv1.ListPaymentsBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertPayments(items)
	if err != nil {
		return nil, err
	}
	return &paymentsv1.PaymentListResponse{Data: data, Meta: converter.ConvertMeta(meta)}, nil
}

func (a *api) PatchPayment(ctx context.Context, req *paymentsv1.PaymentPatchRequest, params paymentsv1.PatchPaymentParams) (paymentsv1.PatchPaymentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := paymentsv1.PatchPaymentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.paymentService.Update(ctx, schoolID, params.PaymentId.String(), converter.ConvertPaymentPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := paymentsv1.PatchPaymentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := paymentsv1.PatchPaymentNotFound(newErrorBody(ctx, "not_found", "payment not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := paymentsv1.PatchPaymentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertPayment(item)
	if err != nil {
		return nil, err
	}
	return &paymentsv1.PaymentResponse{Data: data}, nil
}

func (a *api) DeletePayment(ctx context.Context, params paymentsv1.DeletePaymentParams) (paymentsv1.DeletePaymentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := paymentsv1.DeletePaymentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.paymentService.Delete(ctx, schoolID, params.PaymentId.String()); err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := paymentsv1.DeletePaymentNotFound(newErrorBody(ctx, "not_found", "payment not found"))
			return &res, nil
		case errors.Is(err, svc.ErrInvalidArgument):
			res := paymentsv1.DeletePaymentForbidden(newErrorBody(ctx, "forbidden", "forbidden"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &paymentsv1.DeletePaymentNoContent{}, nil
}

func (a *api) GetStudentBalance(ctx context.Context, params paymentsv1.GetStudentBalanceParams) (paymentsv1.GetStudentBalanceRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := paymentsv1.GetStudentBalanceUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.paymentService.GetBalance(ctx, schoolID, params.StudentId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := paymentsv1.GetStudentBalanceNotFound(newErrorBody(ctx, "not_found", "student not found"))
			return &res, nil
		case errors.Is(err, svc.ErrInvalidArgument):
			res := paymentsv1.GetStudentBalanceUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertBalance(item)
	if err != nil {
		return nil, err
	}
	return &paymentsv1.StudentBalanceResponse{Data: data}, nil
}

func (a *api) UpsertStudentContract(ctx context.Context, req *paymentsv1.StudentContractUpsertRequest, params paymentsv1.UpsertStudentContractParams) (paymentsv1.UpsertStudentContractRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := paymentsv1.UpsertStudentContractUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.paymentService.UpsertContract(ctx, schoolID, params.StudentId.String(), decimal.NewFromFloat(req.ContractTotal))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := paymentsv1.UpsertStudentContractBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := paymentsv1.UpsertStudentContractNotFound(newErrorBody(ctx, "not_found", "student not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := paymentsv1.UpsertStudentContractConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertBalance(item)
	if err != nil {
		return nil, err
	}
	return &paymentsv1.StudentBalanceResponse{Data: data}, nil
}
