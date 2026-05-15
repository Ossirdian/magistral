package lesson

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/lesson/internal/model"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
)

func (s *Service) List(ctx context.Context, schoolID string, filter model.LessonFilter) ([]model.Lesson, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.Status != nil && !isValidStatus(*filter.Status) {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo) {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}

	filter.BranchID, _ = normalizeOptionalID(filter.BranchID)
	normalizeID := func(v **string) {
		if *v == nil {
			return
		}
		value := strings.TrimSpace(**v)
		if value == "" {
			*v = nil
			return
		}
		*v = &value
	}
	normalizeID(&filter.StudentID)
	normalizeID(&filter.InstructorID)
	normalizeID(&filter.CarID)

	filter = normalizeFilter(filter)
	return s.repo.List(ctx, schoolID, filter)
}
