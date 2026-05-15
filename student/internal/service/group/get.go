package group

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Get(ctx context.Context, schoolID, groupID string) (model.Group, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(groupID) == "" {
		return model.Group{}, svc.ErrInvalidArgument
	}

	return s.repo.GetByID(ctx, schoolID, groupID)
}
