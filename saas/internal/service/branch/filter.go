package branch

import "github.com/ossirdian/magistral/saas/internal/model"

const defaultListLimit = 50

func normalizeBranchFilter(filter model.BranchFilter) model.BranchFilter {
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return filter
}
