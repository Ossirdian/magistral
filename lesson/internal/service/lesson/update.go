package lesson

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/lesson/internal/model"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
)

func (s *Service) Update(ctx context.Context, schoolID, lessonID string, patch model.LessonPatch) (model.Lesson, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(lessonID) == "" {
		return model.Lesson{}, svc.ErrInvalidArgument
	}

	normalizeID := func(v **string) error {
		if *v == nil {
			return nil
		}
		value := strings.TrimSpace(**v)
		if value == "" {
			return svc.ErrInvalidArgument
		}
		*v = &value
		return nil
	}

	if patch.BranchIDSet {
		patch.BranchID, _ = normalizeOptionalID(patch.BranchID)
	}
	if err := normalizeID(&patch.StudentID); err != nil {
		return model.Lesson{}, err
	}
	if err := normalizeID(&patch.InstructorID); err != nil {
		return model.Lesson{}, err
	}
	if err := normalizeID(&patch.CarID); err != nil {
		return model.Lesson{}, err
	}
	if patch.Status != nil && !isValidStatus(*patch.Status) {
		return model.Lesson{}, svc.ErrInvalidArgument
	}
	if patch.CommentSet && patch.Comment != nil {
		value := strings.TrimSpace(*patch.Comment)
		patch.Comment = &value
	}

	current, err := s.repo.GetByID(ctx, schoolID, lessonID)
	if err != nil {
		return model.Lesson{}, err
	}

	startAt := current.StartAt
	endAt := current.EndAt
	if patch.StartAt != nil {
		startAt = *patch.StartAt
	}
	if patch.EndAt != nil {
		endAt = *patch.EndAt
	}
	if !validTimeRange(startAt, endAt) {
		return model.Lesson{}, svc.ErrInvalidArgument
	}

	return s.repo.Update(ctx, schoolID, lessonID, patch)
}
