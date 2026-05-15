package model

import "time"

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
	IsGenerated  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LessonCreate struct {
	BranchID     *string
	StudentID    string
	InstructorID string
	CarID        string
	StartAt      time.Time
	EndAt        time.Time
	Status       *LessonStatus
	Comment      *string
	IsGenerated  bool
}

type LessonPatch struct {
	BranchID     *string
	BranchIDSet  bool
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

type BatchReplaceMode string

const (
	BatchReplaceModeAppend               BatchReplaceMode = "append"
	BatchReplaceModeMerge                BatchReplaceMode = "merge"
	BatchReplaceModeReplaceGeneratedOnly BatchReplaceMode = "replace_generated_only"
)

type BatchLessonUpsert struct {
	BranchID     *string
	StudentID    string
	InstructorID string
	CarID        string
	StartAt      time.Time
	EndAt        time.Time
	Comment      *string
}

type BatchApplyRequest struct {
	Mode    BatchReplaceMode
	Lessons []BatchLessonUpsert
}

type BatchConflict struct {
	Index   int
	Code    string
	Message string
}

type BatchApplyResult struct {
	Lessons    []Lesson
	Conflicts  []BatchConflict
	Applied    int
	Requested  int
	Mode       BatchReplaceMode
	Successful bool
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}

type LessonStatus string

const (
	LessonStatusScheduled LessonStatus = "scheduled"
	LessonStatusCompleted LessonStatus = "completed"
	LessonStatusCancelled LessonStatus = "cancelled"
	LessonStatusMissed    LessonStatus = "missed"
)
