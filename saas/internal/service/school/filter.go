package school

import "github.com/ossirdian/magistral/saas/internal/model"

const defaultListLimit = 50

func normalizeSchoolFilter(filter model.SchoolFilter) model.SchoolFilter {
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return filter
}
