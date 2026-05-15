package student

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (s *Service) Get(ctx context.Context, schoolID, studentID string) (model.Student, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(studentID) == "" {
		return model.Student{}, svc.ErrInvalidArgument
	}
	item, err := s.repo.GetByID(ctx, schoolID, studentID)
	if err != nil {
		return model.Student{}, err
	}
	return s.enrichStudent(ctx, schoolID, item)
}
