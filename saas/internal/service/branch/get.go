package branch

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) Get(ctx context.Context, branchID string) (model.Branch, error) {
	return s.branchRepo.GetByID(ctx, branchID)
}
