package student

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) List(ctx context.Context, schoolID string, filter model.StudentFilter) ([]model.Student, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.Status != nil && !isValidStatus(*filter.Status) {
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
	if filter.GroupID != nil {
		value := strings.TrimSpace(*filter.GroupID)
		if value == "" {
			filter.GroupID = nil
		} else {
			filter.GroupID = &value
		}
	}

	filter = normalizeFilter(filter)
	items, meta, err := s.repo.List(ctx, schoolID, filter)
	if err != nil {
		return nil, model.PageMeta{}, err
	}
	items, err = s.enrichStudents(ctx, schoolID, items)
	if err != nil {
		return nil, model.PageMeta{}, err
	}
	return items, meta, nil
}
