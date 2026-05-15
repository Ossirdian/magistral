package student

import (
	"context"
	"strings"

	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Delete(ctx context.Context, schoolID, studentID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(studentID) == "" {
		return svc.ErrInvalidArgument
	}
	return s.repo.Delete(ctx, schoolID, studentID)
}
