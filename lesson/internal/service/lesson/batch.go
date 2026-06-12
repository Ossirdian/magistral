package lesson

import (
	"context"
	"strings"
	"time"

	"github.com/ossirdian/magistral/lesson/internal/model"
	repo "github.com/ossirdian/magistral/lesson/internal/repository"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
	ptracing "github.com/ossirdian/magistral/platform/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) ApplyBatch(ctx context.Context, schoolID string, input model.BatchApplyRequest) (model.BatchApplyResult, error) {
	ctx, span := ptracing.StartSpan(ctx, "lesson.apply_batch",
		trace.WithAttributes(
			attribute.String("lesson.school_id", schoolID),
			attribute.String("lesson.batch.mode", string(input.Mode)),
			attribute.Int("lesson.batch.requested", len(input.Lessons)),
		),
	)
	defer span.End()

	if strings.TrimSpace(schoolID) == "" || !isValidBatchMode(input.Mode) || len(input.Lessons) == 0 {
		span.SetAttributes(attribute.Bool("lesson.batch.invalid", true))
		return model.BatchApplyResult{}, svc.ErrInvalidArgument
	}

	normalized := make([]model.BatchLessonUpsert, 0, len(input.Lessons))
	studentSet := make(map[string]struct{}, len(input.Lessons))
	branchSet := make(map[string]struct{}, len(input.Lessons))
	includeNullBranch := false
	var dateFrom, dateTo time.Time

	for idx, item := range input.Lessons {
		branchID, _ := normalizeOptionalID(item.BranchID)
		item.BranchID = branchID
		item.StudentID = strings.TrimSpace(item.StudentID)
		item.InstructorID = strings.TrimSpace(item.InstructorID)
		item.CarID = strings.TrimSpace(item.CarID)
		if item.StudentID == "" || item.InstructorID == "" || item.CarID == "" || !validTimeRange(item.StartAt, item.EndAt) {
			span.SetAttributes(attribute.Bool("lesson.batch.invalid", true))
			return model.BatchApplyResult{}, svc.ErrInvalidArgument
		}
		if item.Comment != nil {
			comment := strings.TrimSpace(*item.Comment)
			item.Comment = &comment
		}
		if idx == 0 || item.StartAt.Before(dateFrom) {
			dateFrom = item.StartAt
		}
		if idx == 0 || item.EndAt.After(dateTo) {
			dateTo = item.EndAt
		}
		studentSet[item.StudentID] = struct{}{}
		if item.BranchID == nil {
			includeNullBranch = true
		} else {
			branchSet[*item.BranchID] = struct{}{}
		}
		normalized = append(normalized, item)
	}

	if input.Mode == model.BatchReplaceModeReplaceGeneratedOnly {
		branchIDs := make([]string, 0, len(branchSet))
		for branchID := range branchSet {
			branchIDs = append(branchIDs, branchID)
		}
		if err := s.repo.DeleteGeneratedInWindowByBranches(ctx, schoolID, branchIDs, includeNullBranch, dateFrom, dateTo); err != nil {
			span.RecordError(err)
			return model.BatchApplyResult{}, err
		}
	}

	if input.Mode == model.BatchReplaceModeMerge {
		studentIDs := make([]string, 0, len(studentSet))
		for studentID := range studentSet {
			studentIDs = append(studentIDs, studentID)
		}
		if err := s.repo.DeleteGeneratedInWindow(ctx, schoolID, studentIDs, dateFrom, dateTo); err != nil {
			span.RecordError(err)
			return model.BatchApplyResult{}, err
		}
	}

	result := model.BatchApplyResult{
		Requested: len(normalized),
		Mode:      input.Mode,
	}
	for idx, item := range normalized {
		createCtx, createSpan := ptracing.StartSpan(ctx, "lesson.apply_batch.create_lesson",
			trace.WithAttributes(
				attribute.Int("lesson.batch.index", idx),
				attribute.String("lesson.student_id", item.StudentID),
				attribute.String("lesson.instructor_id", item.InstructorID),
				attribute.String("lesson.car_id", item.CarID),
			),
		)
		created, err := s.repo.Create(createCtx, schoolID, model.LessonCreate{
			BranchID:     item.BranchID,
			StudentID:    item.StudentID,
			InstructorID: item.InstructorID,
			CarID:        item.CarID,
			StartAt:      item.StartAt,
			EndAt:        item.EndAt,
			Comment:      item.Comment,
			IsGenerated:  true,
			Status:       ptrStatus(model.LessonStatusScheduled),
		})
		if err != nil {
			createSpan.RecordError(err)
			createSpan.End()
			span.RecordError(err)
			code := "internal_error"
			message := "failed to apply lesson"
			if err == repo.ErrConflict {
				code = "conflict"
				message = "lesson conflicts with schedule"
			}
			result.Conflicts = append(result.Conflicts, model.BatchConflict{
				Index:   idx,
				Code:    code,
				Message: message,
			})
			continue
		}
		createSpan.SetAttributes(attribute.String("lesson.id", created.ID))
		createSpan.End()
		result.Lessons = append(result.Lessons, created)
		result.Applied++
	}
	result.Successful = len(result.Conflicts) == 0
	span.SetAttributes(
		attribute.Int("lesson.batch.applied", result.Applied),
		attribute.Int("lesson.batch.conflicts", len(result.Conflicts)),
		attribute.Bool("lesson.batch.successful", result.Successful),
	)
	return result, nil
}

func ptrStatus(status model.LessonStatus) *model.LessonStatus {
	return &status
}
