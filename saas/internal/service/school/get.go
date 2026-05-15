package school

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) Get(ctx context.Context, schoolID string) (model.School, error) {
	return s.schoolRepo.GetByID(ctx, schoolID)
}
