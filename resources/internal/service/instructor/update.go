package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) Update(ctx context.Context, schoolID, instructorID string, patch model.InstructorPatch) (model.Instructor, error) {
	return s.instructorRepo.Update(ctx, schoolID, instructorID, patch)
}
