package branch

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) List(ctx context.Context, filter model.BranchFilter) ([]model.Branch, model.PageMeta, error) {
	filter = normalizeBranchFilter(filter)
	return s.branchRepo.List(ctx, filter)
}
