package lesson

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/lesson/internal/model"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
)

func (s *Service) Create(ctx context.Context, schoolID string, input model.LessonCreate) (model.Lesson, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.Lesson{}, svc.ErrInvalidArgument
	}

	branchID, _ := normalizeOptionalID(input.BranchID)
	input.BranchID = branchID
	input.StudentID = strings.TrimSpace(input.StudentID)
	input.InstructorID = strings.TrimSpace(input.InstructorID)
	input.CarID = strings.TrimSpace(input.CarID)
	if input.StudentID == "" || input.InstructorID == "" || input.CarID == "" {
		return model.Lesson{}, svc.ErrInvalidArgument
	}

	if input.Status == nil {
		defaultStatus := model.LessonStatusScheduled
		input.Status = &defaultStatus
	}
	if !isValidStatus(*input.Status) || !validTimeRange(input.StartAt, input.EndAt) {
		return model.Lesson{}, svc.ErrInvalidArgument
	}
	if input.Comment != nil {
		comment := strings.TrimSpace(*input.Comment)
		input.Comment = &comment
	}

	return s.repo.Create(ctx, schoolID, input)
}
