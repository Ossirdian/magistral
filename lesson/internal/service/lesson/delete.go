package lesson

import (
	"context"
	"strings"

	svc "github.com/ossirdian/magistral/lesson/internal/service"
)

func (s *Service) Delete(ctx context.Context, schoolID, lessonID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(lessonID) == "" {
		return svc.ErrInvalidArgument
	}

	return s.repo.Delete(ctx, schoolID, lessonID)
}
