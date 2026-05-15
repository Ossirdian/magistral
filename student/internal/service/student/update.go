package student

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Update(ctx context.Context, schoolID, studentID string, patch model.StudentPatch) (model.Student, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(studentID) == "" {
		return model.Student{}, svc.ErrInvalidArgument
	}

	if patch.BranchID != nil {
		value := strings.TrimSpace(*patch.BranchID)
		if value == "" {
			return model.Student{}, svc.ErrInvalidArgument
		}
		patch.BranchID = &value
	}
	if patch.GroupIDSet && patch.GroupID != nil {
		value := strings.TrimSpace(*patch.GroupID)
		if value == "" {
			patch.GroupID = nil
		} else {
			patch.GroupID = &value
		}
	}
	if patch.ProgramID != nil {
		value := strings.TrimSpace(*patch.ProgramID)
		if value == "" {
			return model.Student{}, svc.ErrInvalidArgument
		}
		if _, err := s.programs.GetByID(ctx, schoolID, value); err != nil {
			return model.Student{}, err
		}
		patch.ProgramID = &value
	}
	if patch.FullName != nil {
		value := strings.TrimSpace(*patch.FullName)
		if value == "" {
			return model.Student{}, svc.ErrInvalidArgument
		}
		patch.FullName = &value
	}
	if patch.Phone != nil {
		value := strings.TrimSpace(*patch.Phone)
		if value == "" {
			return model.Student{}, svc.ErrInvalidArgument
		}
		patch.Phone = &value
	}
	if patch.Status != nil && !isValidStatus(*patch.Status) {
		return model.Student{}, svc.ErrInvalidArgument
	}
	if patch.ManualHoursCompleted != nil && *patch.ManualHoursCompleted < 0 {
		return model.Student{}, svc.ErrInvalidArgument
	}
	if !areValidPreferredWeekdays(patch.PreferredWeekdays) || !areValidPreferredTimeWindows(patch.PreferredTimeWindows) {
		return model.Student{}, svc.ErrInvalidArgument
	}

	item, err := s.repo.Update(ctx, schoolID, studentID, patch)
	if err != nil {
		return model.Student{}, err
	}
	return s.enrichStudent(ctx, schoolID, item)
}
