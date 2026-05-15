package student

import "github.com/ossirdian/magistral/student/internal/model"

func isValidCreate(input model.StudentCreate) bool {
	if input.BranchID == "" || input.ProgramID == "" || input.FullName == "" || input.Phone == "" || input.Status == nil {
		return false
	}
	if input.ManualHoursCompleted != nil && *input.ManualHoursCompleted < 0 {
		return false
	}
	if !areValidPreferredWeekdays(input.PreferredWeekdays) || !areValidPreferredTimeWindows(input.PreferredTimeWindows) {
		return false
	}
	return isValidStatus(*input.Status)
}

func isValidStatus(status model.StudentStatus) bool {
	switch status {
	case model.StudentStatusNew, model.StudentStatusActive, model.StudentStatusPaused, model.StudentStatusGraduated, model.StudentStatusArchived:
		return true
	default:
		return false
	}
}

func areValidPreferredWeekdays(values []model.StudentPreferredWeekday) bool {
	for _, value := range values {
		switch value {
		case model.StudentPreferredWeekdayMonday,
			model.StudentPreferredWeekdayTuesday,
			model.StudentPreferredWeekdayWednesday,
			model.StudentPreferredWeekdayThursday,
			model.StudentPreferredWeekdayFriday,
			model.StudentPreferredWeekdaySaturday,
			model.StudentPreferredWeekdaySunday:
		default:
			return false
		}
	}
	return true
}

func areValidPreferredTimeWindows(values []model.StudentPreferredTimeWindow) bool {
	for _, value := range values {
		switch value {
		case model.StudentPreferredTimeWindowMorning,
			model.StudentPreferredTimeWindowDaytime,
			model.StudentPreferredTimeWindowEvening:
		default:
			return false
		}
	}
	return true
}
