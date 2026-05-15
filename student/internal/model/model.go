package model

import "time"

type Student struct {
	ID                   string
	SchoolID             string
	BranchID             string
	GroupID              *string
	ProgramID            *string
	FullName             string
	Phone                string
	Status               StudentStatus
	PreferredWeekdays    []StudentPreferredWeekday
	PreferredTimeWindows []StudentPreferredTimeWindow
	ManualHoursCompleted float64
	LessonHoursCompleted float64
	TotalHoursCompleted  float64
	RemainingHours       float64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type StudentCreate struct {
	BranchID             string
	GroupID              *string
	ProgramID            string
	FullName             string
	Phone                string
	Status               *StudentStatus
	PreferredWeekdays    []StudentPreferredWeekday
	PreferredTimeWindows []StudentPreferredTimeWindow
	ManualHoursCompleted *float64
}

type StudentPatch struct {
	BranchID             *string
	GroupID              *string
	GroupIDSet           bool
	ProgramID            *string
	FullName             *string
	Phone                *string
	Status               *StudentStatus
	PreferredWeekdays    []StudentPreferredWeekday
	PreferredTimeWindows []StudentPreferredTimeWindow
	ManualHoursCompleted *float64
}

type StudentFilter struct {
	BranchID *string
	GroupID  *string
	Status   *StudentStatus
	Search   *string
	Limit    int
	Offset   int
}

type Group struct {
	ID        string
	SchoolID  string
	BranchID  string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GroupCreate struct {
	BranchID string
	Name     string
}

type GroupPatch struct {
	BranchID *string
	Name     *string
}

type GroupFilter struct {
	BranchID *string
	Search   *string
	Limit    int
	Offset   int
}

type StudyProgram struct {
	ID           string
	SchoolID     string
	Name         string
	Transmission TransmissionType
	TotalHours   float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StudyProgramCreate struct {
	Name         string
	Transmission TransmissionType
	TotalHours   float64
}

type StudyProgramPatch struct {
	Name         *string
	Transmission *TransmissionType
	TotalHours   *float64
}

type StudyProgramFilter struct {
	Search string
	Limit  int
	Offset int
}

type Lesson struct {
	ID           string
	SchoolID     string
	BranchID     string
	StudentID    string
	InstructorID string
	CarID        string
	StartAt      time.Time
	EndAt        time.Time
	Status       LessonStatus
	Comment      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LessonCreate struct {
	BranchID     string
	StudentID    string
	InstructorID string
	CarID        string
	StartAt      time.Time
	EndAt        time.Time
	Status       *LessonStatus
	Comment      *string
}

type LessonPatch struct {
	BranchID     *string
	StudentID    *string
	InstructorID *string
	CarID        *string
	StartAt      *time.Time
	EndAt        *time.Time
	Status       *LessonStatus
	Comment      *string
	CommentSet   bool
}

type LessonFilter struct {
	BranchID     *string
	StudentID    *string
	InstructorID *string
	CarID        *string
	Status       *LessonStatus
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	Offset       int
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}

type PlanningRequestCreate struct {
	PeriodStart     time.Time
	PeriodEnd       time.Time
	StudentIDs      []string
	GroupIDs        []string
	BranchIDs       []string
	ReplaceStrategy PlanningReplaceStrategy
	RequestID       *string
}

type PlanningRequestMessage struct {
	SchoolID           string                      `json:"school_id"`
	PeriodStart        time.Time                   `json:"period_start"`
	PeriodEnd          time.Time                   `json:"period_end"`
	StudentIDs         []string                    `json:"student_ids"`
	StudentPreferences []PlanningStudentPreference `json:"student_preferences,omitempty"`
	GroupIDs           []string                    `json:"group_ids"`
	BranchIDs          []string                    `json:"branch_ids"`
	ReplaceStrategy    PlanningReplaceStrategy     `json:"replace_strategy"`
	RequestedByUserID  string                      `json:"requested_by_user_id"`
	RequestID          string                      `json:"request_id"`
}

type PlanningRequestAccepted struct {
	RequestID string
	Status    PlanningRequestStatus
}

type PlanningStudentPreference struct {
	StudentID            string                       `json:"student_id"`
	PreferredWeekdays    []StudentPreferredWeekday    `json:"preferred_weekdays,omitempty"`
	PreferredTimeWindows []StudentPreferredTimeWindow `json:"preferred_time_windows,omitempty"`
}

type (
	StudentStatus              string
	TransmissionType           string
	LessonStatus               string
	PlanningReplaceStrategy    string
	PlanningRequestStatus      string
	StudentPreferredWeekday    string
	StudentPreferredTimeWindow string
)

const (
	StudentStatusNew       StudentStatus = "new"
	StudentStatusActive    StudentStatus = "active"
	StudentStatusPaused    StudentStatus = "paused"
	StudentStatusGraduated StudentStatus = "graduated"
	StudentStatusArchived  StudentStatus = "archived"

	StudentPreferredWeekdayMonday    StudentPreferredWeekday = "monday"
	StudentPreferredWeekdayTuesday   StudentPreferredWeekday = "tuesday"
	StudentPreferredWeekdayWednesday StudentPreferredWeekday = "wednesday"
	StudentPreferredWeekdayThursday  StudentPreferredWeekday = "thursday"
	StudentPreferredWeekdayFriday    StudentPreferredWeekday = "friday"
	StudentPreferredWeekdaySaturday  StudentPreferredWeekday = "saturday"
	StudentPreferredWeekdaySunday    StudentPreferredWeekday = "sunday"

	StudentPreferredTimeWindowMorning StudentPreferredTimeWindow = "morning"
	StudentPreferredTimeWindowDaytime StudentPreferredTimeWindow = "daytime"
	StudentPreferredTimeWindowEvening StudentPreferredTimeWindow = "evening"

	TransmissionTypeManual    TransmissionType = "manual"
	TransmissionTypeAutomatic TransmissionType = "automatic"

	LessonStatusScheduled LessonStatus = "scheduled"
	LessonStatusCompleted LessonStatus = "completed"
	LessonStatusCancelled LessonStatus = "cancelled"
	LessonStatusMissed    LessonStatus = "missed"

	PlanningReplaceStrategyAppend               PlanningReplaceStrategy = "append"
	PlanningReplaceStrategyMerge                PlanningReplaceStrategy = "merge"
	PlanningReplaceStrategyReplaceGeneratedOnly PlanningReplaceStrategy = "replace_generated_only"

	PlanningRequestStatusQueued PlanningRequestStatus = "queued"
)
