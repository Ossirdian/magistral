package student

import (
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
)

const defaultListLimit = 50

func normalizeFilter(filter model.StudentFilter) model.StudentFilter {
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.Search != nil {
		value := strings.TrimSpace(*filter.Search)
		if value == "" {
			filter.Search = nil
		} else {
			filter.Search = &value
		}
	}
	return filter
}
