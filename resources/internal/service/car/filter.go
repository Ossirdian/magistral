package car

import "github.com/ossirdian/magistral/resources/internal/model"

const defaultListLimit = 50

func normalizeCarFilter(filter model.CarFilter) model.CarFilter {
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return filter
}
