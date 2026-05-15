package lesson

import (
	"strings"
	"time"

	"github.com/ossirdian/magistral/lesson/internal/model"
)

func isValidStatus(status model.LessonStatus) bool {
	switch status {
	case model.LessonStatusScheduled, model.LessonStatusCompleted, model.LessonStatusCancelled, model.LessonStatusMissed:
		return true
	default:
		return false
	}
}

func isValidBatchMode(mode model.BatchReplaceMode) bool {
	switch mode {
	case model.BatchReplaceModeAppend, model.BatchReplaceModeMerge, model.BatchReplaceModeReplaceGeneratedOnly:
		return true
	default:
		return false
	}
}

func validTimeRange(startAt, endAt time.Time) bool {
	return startAt.Before(endAt)
}

func normalizeOptionalID(value *string) (*string, bool) {
	if value == nil {
		return nil, true
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, true
	}
	return &trimmed, true
}
