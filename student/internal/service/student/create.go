package student

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Create(ctx context.Context, schoolID string, input model.StudentCreate) (model.Student, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.Student{}, svc.ErrInvalidArgument
	}

	input.BranchID = strings.TrimSpace(input.BranchID)
	if input.GroupID != nil {
		value := strings.TrimSpace(*input.GroupID)
		if value == "" {
			input.GroupID = nil
		} else {
			input.GroupID = &value
		}
	}
	input.ProgramID = strings.TrimSpace(input.ProgramID)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Status == nil {
		defaultStatus := model.StudentStatusNew
		input.Status = &defaultStatus
	}

	if !isValidCreate(input) {
		return model.Student{}, svc.ErrInvalidArgument
	}
	if _, err := s.programs.GetByID(ctx, schoolID, input.ProgramID); err != nil {
		return model.Student{}, err
	}

	item, err := s.repo.Create(ctx, schoolID, input)
	if err != nil {
		return model.Student{}, err
	}
	return s.enrichStudent(ctx, schoolID, item)
}
