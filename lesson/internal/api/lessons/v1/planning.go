package api

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/lesson/internal/httpauth"
	"github.com/ossirdian/magistral/lesson/internal/planning"
	lessonsv1 "github.com/ossirdian/magistral/shared/pkg/lessons/v1"
)

func (a *api) CreatePlanningRequest(ctx context.Context, req *lessonsv1.PlanningRequestCreateRequest) (lessonsv1.CreatePlanningRequestRes, error) {
	schoolID, ok := httpauth.SchoolIDFromContext(ctx)
	if !ok {
		res := lessonsv1.CreatePlanningRequestUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}
	userID, _ := httpauth.UserIDFromContext(ctx)

	accepted, err := a.planningProducer.Publish(ctx, schoolID, userID, planning.RequestCreate{
		PeriodStart:     req.PeriodStart,
		PeriodEnd:       req.PeriodEnd,
		StudentIDs:      uuidStrings(req.StudentIds),
		GroupIDs:        uuidStrings(req.GroupIds),
		BranchIDs:       uuidStrings(req.BranchIds),
		ReplaceStrategy: string(req.ReplaceStrategy),
		RequestID:       optNilStringToPtr(req.RequestID),
	})
	if err != nil {
		switch {
		case errors.Is(err, planning.ErrInvalidRequest):
			res := lessonsv1.CreatePlanningRequestBadRequest(newErrorBody(ctx, "bad_request", "invalid planning request"))
			return &res, nil
		default:
			res := lessonsv1.CreatePlanningRequestServiceUnavailable(newErrorBody(ctx, "service_unavailable", "failed to queue planning request"))
			return &res, nil
		}
	}
	if a.planningStatuses != nil {
		a.planningStatuses.SaveQueued(schoolID, planning.RequestCreate{
			PeriodStart:     req.PeriodStart,
			PeriodEnd:       req.PeriodEnd,
			StudentIDs:      uuidStrings(req.StudentIds),
			GroupIDs:        uuidStrings(req.GroupIds),
			BranchIDs:       uuidStrings(req.BranchIds),
			ReplaceStrategy: string(req.ReplaceStrategy),
			RequestID:       optNilStringToPtr(req.RequestID),
		}, planning.RequestAccepted{RequestID: accepted.RequestID, Status: accepted.Status})
	}

	return &lessonsv1.PlanningRequestResponse{
		Data: lessonsv1.PlanningRequestAccepted{
			RequestID: accepted.RequestID,
			Status:    accepted.Status,
		},
	}, nil
}

func uuidStrings(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}

func optNilStringToPtr(value lessonsv1.OptNilString) *string {
	if !value.IsSet() || value.IsNull() {
		return nil
	}
	v, ok := value.Get()
	if !ok {
		return nil
	}
	return &v
}
