package group

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Update(ctx context.Context, schoolID, groupID string, patch model.GroupPatch) (model.Group, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(groupID) == "" {
		return model.Group{}, svc.ErrInvalidArgument
	}

	if patch.BranchID != nil {
		value := strings.TrimSpace(*patch.BranchID)
		if value == "" {
			return model.Group{}, svc.ErrInvalidArgument
		}
		patch.BranchID = &value
	}
	if patch.Name != nil {
		value := strings.TrimSpace(*patch.Name)
		if value == "" {
			return model.Group{}, svc.ErrInvalidArgument
		}
		patch.Name = &value
	}

	return s.repo.Update(ctx, schoolID, groupID, patch)
}
