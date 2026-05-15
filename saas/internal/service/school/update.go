package school

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) Update(ctx context.Context, schoolID string, patch model.SchoolPatch) (model.School, error) {
	return s.schoolRepo.Update(ctx, schoolID, patch)
}
