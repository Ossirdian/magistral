package lesson

import "github.com/ossirdian/magistral/lesson/internal/model"

const defaultListLimit = 50

func normalizeFilter(filter model.LessonFilter) model.LessonFilter {
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return filter
}
