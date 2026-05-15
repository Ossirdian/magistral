package group

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) List(ctx context.Context, schoolID string, filter model.GroupFilter) ([]model.Group, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.BranchID != nil {
		value := strings.TrimSpace(*filter.BranchID)
		if value == "" {
			filter.BranchID = nil
		} else {
			filter.BranchID = &value
		}
	}

	filter = normalizeFilter(filter)
	return s.repo.List(ctx, schoolID, filter)
}
