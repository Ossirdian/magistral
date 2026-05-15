package branch

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) Create(ctx context.Context, input model.BranchCreate) (model.Branch, error) {
	return s.branchRepo.Create(ctx, input)
}
