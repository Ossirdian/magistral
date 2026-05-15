package group

import (
	"context"
	"strings"

	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Delete(ctx context.Context, schoolID, groupID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(groupID) == "" {
		return svc.ErrInvalidArgument
	}

	return s.repo.Delete(ctx, schoolID, groupID)
}
