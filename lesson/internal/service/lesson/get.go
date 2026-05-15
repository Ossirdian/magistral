package lesson

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/lesson/internal/model"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
)

func (s *Service) Get(ctx context.Context, schoolID, lessonID string) (model.Lesson, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(lessonID) == "" {
		return model.Lesson{}, svc.ErrInvalidArgument
	}

	return s.repo.GetByID(ctx, schoolID, lessonID)
}
