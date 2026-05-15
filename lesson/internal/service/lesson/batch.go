package lesson

import (
	"context"
	"strings"
	"time"

	"github.com/ossirdian/magistral/lesson/internal/model"
	repo "github.com/ossirdian/magistral/lesson/internal/repository"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
)

func (s *Service) ApplyBatch(ctx context.Context, schoolID string, input model.BatchApplyRequest) (model.BatchApplyResult, error) {
	if strings.TrimSpace(schoolID) == "" || !isValidBatchMode(input.Mode) || len(input.Lessons) == 0 {
		return model.BatchApplyResult{}, svc.ErrInvalidArgument
	}

	normalized := make([]model.BatchLessonUpsert, 0, len(input.Lessons))
	studentSet := make(map[string]struct{}, len(input.Lessons))
	var dateFrom, dateTo time.Time

	for idx, item := range input.Lessons {
		branchID, _ := normalizeOptionalID(item.BranchID)
		item.BranchID = branchID
		item.StudentID = strings.TrimSpace(item.StudentID)
		item.InstructorID = strings.TrimSpace(item.InstructorID)
		item.CarID = strings.TrimSpace(item.CarID)
		if item.StudentID == "" || item.InstructorID == "" || item.CarID == "" || !validTimeRange(item.StartAt, item.EndAt) {
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
		normalized = append(normalized, item)
	}

	if input.Mode == model.BatchReplaceModeMerge || input.Mode == model.BatchReplaceModeReplaceGeneratedOnly {
		studentIDs := make([]string, 0, len(studentSet))
		for studentID := range studentSet {
			studentIDs = append(studentIDs, studentID)
		}
		if err := s.repo.DeleteGeneratedInWindow(ctx, schoolID, studentIDs, dateFrom, dateTo); err != nil {
			return model.BatchApplyResult{}, err
		}
	}

	result := model.BatchApplyResult{
		Requested: len(normalized),
		Mode:      input.Mode,
	}
	for idx, item := range normalized {
		created, err := s.repo.Create(ctx, schoolID, model.LessonCreate{
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
		result.Lessons = append(result.Lessons, created)
		result.Applied++
	}
	result.Successful = len(result.Conflicts) == 0
	return result, nil
}

func ptrStatus(status model.LessonStatus) *model.LessonStatus {
	return &status
}
