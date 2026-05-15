package school

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) Create(ctx context.Context, input model.SchoolCreate) (model.School, error) {
	return s.schoolRepo.Create(ctx, input)
}
