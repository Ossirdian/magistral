package student

import "github.com/ossirdian/magistral/student/internal/model"

func preferredWeekdaysToStrings(values []model.StudentPreferredWeekday) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func preferredTimeWindowsToStrings(values []model.StudentPreferredTimeWindow) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}
