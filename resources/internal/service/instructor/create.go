package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) Create(ctx context.Context, schoolID string, input model.InstructorCreate) (model.Instructor, error) {
	return s.instructorRepo.Create(ctx, schoolID, input)
}
