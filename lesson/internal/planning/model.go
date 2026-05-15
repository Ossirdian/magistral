package planning

import "time"

type RequestCreate struct {
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	StudentIDs      []string  `json:"student_ids"`
	GroupIDs        []string  `json:"group_ids"`
	BranchIDs       []string  `json:"branch_ids"`
	ReplaceStrategy string    `json:"replace_strategy"`
	RequestID       *string   `json:"request_id"`
}

type RequestMessage struct {
	SchoolID           string                       `json:"school_id"`
	PeriodStart        time.Time                    `json:"period_start"`
	PeriodEnd          time.Time                    `json:"period_end"`
	StudentIDs         []string                     `json:"student_ids"`
	StudentPreferences map[string]StudentPreference `json:"student_preferences,omitempty"`
	GroupIDs           []string                     `json:"group_ids"`
	BranchIDs          []string                     `json:"branch_ids"`
	ReplaceStrategy    string                       `json:"replace_strategy"`
	RequestedByUserID  string                       `json:"requested_by_user_id"`
	RequestID          string                       `json:"request_id"`
}

type StudentPreference struct {
	PreferredWeekdays    []string `json:"preferred_weekdays,omitempty"`
	PreferredTimeWindows []string `json:"preferred_time_windows,omitempty"`
	Transmission         string   `json:"transmission,omitempty"`
}

type RequestAccepted struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

type Response struct {
	Data RequestAccepted `json:"data"`
}
