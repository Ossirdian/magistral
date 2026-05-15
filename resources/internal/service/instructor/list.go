package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) List(ctx context.Context, schoolID string, filter model.InstructorFilter) ([]model.Instructor, model.PageMeta, error) {
	filter = normalizeInstructorFilter(filter)
	return s.instructorRepo.List(ctx, schoolID, filter)
}
