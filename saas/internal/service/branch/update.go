package branch

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) Update(ctx context.Context, branchID string, patch model.BranchPatch) (model.Branch, error) {
	return s.branchRepo.Update(ctx, branchID, patch)
}
