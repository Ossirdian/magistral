package planning

import (
	"testing"
	"time"

	"github.com/ossirdian/magistral/lesson/internal/model"
)

func TestResultMessageToBatchApplyRequest(t *testing.T) {
	start := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)

	result := ResultMessage{
		RequestID:       "request-1",
		SchoolID:        "school-1",
		ReplaceStrategy: "merge",
		TraceContext: map[string]string{
			"traceparent": "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01",
		},
		Lessons: []ResultLesson{
			{
				StudentID:    "11111111-1111-1111-1111-111111111001",
				InstructorID: "11111111-1111-1111-1111-111111111101",
				CarID:        "11111111-1111-1111-1111-111111111201",
				StartAt:      start,
				EndAt:        end,
			},
		},
	}

	batch := result.toBatchApplyRequest()

	if batch.Mode != model.BatchReplaceModeMerge {
		t.Fatalf("expected merge mode, got %q", batch.Mode)
	}
	if len(batch.Lessons) != 1 {
		t.Fatalf("expected one lesson, got %d", len(batch.Lessons))
	}
	lesson := batch.Lessons[0]
	if lesson.StudentID != "11111111-1111-1111-1111-111111111001" ||
		lesson.InstructorID != "11111111-1111-1111-1111-111111111101" ||
		lesson.CarID != "11111111-1111-1111-1111-111111111201" {
		t.Fatalf("unexpected lesson identifiers: %+v", lesson)
	}
	if lesson.Comment == nil || *lesson.Comment == "" {
		t.Fatal("expected generated planning comment")
	}
}

func TestResultMessageToBatchApplyRequestSkipsInvalidIdentifiers(t *testing.T) {
	start := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)

	result := ResultMessage{
		ReplaceStrategy: "append",
		Lessons: []ResultLesson{
			{
				StudentID:    "not-a-uuid",
				InstructorID: "11111111-1111-1111-1111-111111111101",
				CarID:        "11111111-1111-1111-1111-111111111201",
				StartAt:      start,
				EndAt:        start.Add(90 * time.Minute),
			},
		},
	}

	batch := result.toBatchApplyRequest()

	if len(batch.Lessons) != 0 {
		t.Fatalf("expected invalid lesson to be skipped, got %d lessons", len(batch.Lessons))
	}
}
