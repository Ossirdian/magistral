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
	TraceContext       map[string]string            `json:"trace_context,omitempty"`
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

type RequestStatus struct {
	RequestID           string         `json:"request_id"`
	Status              string         `json:"status"`
	PeriodStart         time.Time      `json:"period_start"`
	PeriodEnd           time.Time      `json:"period_end"`
	StudentIDs          []string       `json:"student_ids"`
	GroupIDs            []string       `json:"group_ids"`
	BranchIDs           []string       `json:"branch_ids"`
	ReplaceStrategy     string         `json:"replace_strategy"`
	QueuedAt            time.Time      `json:"queued_at"`
	CompletedAt         *time.Time     `json:"completed_at,omitempty"`
	ScheduledLessons    int            `json:"scheduled_lessons"`
	AppliedLessons      int            `json:"applied_lessons"`
	Conflicts           int            `json:"conflicts"`
	UnassignedStudents  int            `json:"unassigned_students"`
	Score               float64        `json:"score"`
	ResourceIdleSeconds float64        `json:"resource_idle_seconds"`
	ScoreBreakdown      map[string]any `json:"score_breakdown,omitempty"`
	ApplySuccessful     *bool          `json:"apply_successful,omitempty"`
}

type Response struct {
	Data RequestAccepted `json:"data"`
}

type ResultMessage struct {
	RequestID           string            `json:"request_id"`
	SchoolID            string            `json:"school_id"`
	ReplaceStrategy     string            `json:"replace_strategy"`
	Lessons             []ResultLesson    `json:"lessons"`
	UnassignedStudents  []string          `json:"unassigned_students"`
	Conflicts           []ResultConflict  `json:"conflicts"`
	Score               float64           `json:"score"`
	ResourceIdleSeconds float64           `json:"resource_idle_seconds"`
	ScoreBreakdown      map[string]any    `json:"score_breakdown,omitempty"`
	TraceContext        map[string]string `json:"trace_context,omitempty"`
}

type ResultLesson struct {
	BranchID     *string   `json:"branch_id,omitempty"`
	StudentID    string    `json:"student_id"`
	InstructorID string    `json:"instructor_id"`
	CarID        string    `json:"car_id"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	Status       string    `json:"status"`
}

type ResultConflict struct {
	StudentID string `json:"student_id"`
	Reason    string `json:"reason"`
}
