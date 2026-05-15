package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) Get(ctx context.Context, schoolID, instructorID string) (model.Instructor, error) {
	return s.instructorRepo.GetByID(ctx, schoolID, instructorID)
}
