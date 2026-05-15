package group

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Create(ctx context.Context, schoolID string, input model.GroupCreate) (model.Group, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.Group{}, svc.ErrInvalidArgument
	}

	input.BranchID = strings.TrimSpace(input.BranchID)
	input.Name = strings.TrimSpace(input.Name)
	if input.BranchID == "" || input.Name == "" {
		return model.Group{}, svc.ErrInvalidArgument
	}

	return s.repo.Create(ctx, schoolID, input)
}
