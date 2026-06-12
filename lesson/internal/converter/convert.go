package converter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/lesson/internal/model"
	lessonsv1 "github.com/ossirdian/magistral/shared/pkg/lessons/v1"
)

func ConvertLesson(item model.Lesson) (lessonsv1.Lesson, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return lessonsv1.Lesson{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return lessonsv1.Lesson{}, err
	}
	studentID, err := parseUUID(item.StudentID)
	if err != nil {
		return lessonsv1.Lesson{}, err
	}
	instructorID, err := parseUUID(item.InstructorID)
	if err != nil {
		return lessonsv1.Lesson{}, err
	}
	carID, err := parseUUID(item.CarID)
	if err != nil {
		return lessonsv1.Lesson{}, err
	}

	result := lessonsv1.Lesson{
		ID:           id,
		SchoolID:     schoolID,
		StudentID:    studentID,
		InstructorID: instructorID,
		CarID:        carID,
		StartAt:      item.StartAt,
		EndAt:        item.EndAt,
		Status:       lessonsv1.LessonStatus(item.Status),
		Comment:      ptrStringToOptNilString(item.Comment),
		IsGenerated:  item.IsGenerated,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
	if item.BranchID != "" {
		branchID, err := parseUUID(item.BranchID)
		if err != nil {
			return lessonsv1.Lesson{}, err
		}
		result.BranchID.SetTo(branchID)
	}
	return result, nil
}

func ConvertLessons(items []model.Lesson) ([]lessonsv1.Lesson, error) {
	result := make([]lessonsv1.Lesson, 0, len(items))
	for _, item := range items {
		converted, err := ConvertLesson(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertLessonCreateRequest(req *lessonsv1.LessonCreateRequest) model.LessonCreate {
	return model.LessonCreate{
		BranchID:     optNilUUIDToPtr(req.BranchID),
		StudentID:    req.StudentID.String(),
		InstructorID: req.InstructorID.String(),
		CarID:        req.CarID.String(),
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		Status:       optLessonStatusToPtr(req.Status),
		Comment:      optNilStringToPtr(req.Comment),
	}
}

func ConvertLessonPatchRequest(req *lessonsv1.LessonPatchRequest) model.LessonPatch {
	comment, commentSet := optNilStringToPatch(req.Comment)
	return model.LessonPatch{
		BranchID:     optNilUUIDToPtr(req.BranchID),
		BranchIDSet:  req.BranchID.IsSet(),
		StudentID:    optUUIDToPtr(req.StudentID),
		InstructorID: optUUIDToPtr(req.InstructorID),
		CarID:        optUUIDToPtr(req.CarID),
		StartAt:      optDateTimeToPtr(req.StartAt),
		EndAt:        optDateTimeToPtr(req.EndAt),
		Status:       optLessonStatusToPtr(req.Status),
		Comment:      comment,
		CommentSet:   commentSet,
	}
}

func ConvertLessonFilter(params lessonsv1.ListLessonsParams) model.LessonFilter {
	return model.LessonFilter{
		BranchID:     optUUIDToPtr(params.BranchID),
		StudentID:    optUUIDToPtr(params.StudentID),
		InstructorID: optUUIDToPtr(params.InstructorID),
		CarID:        optUUIDToPtr(params.CarID),
		Status:       optLessonStatusToPtr(params.Status),
		DateFrom:     optDateTimeToPtr(params.DateFrom),
		DateTo:       optDateTimeToPtr(params.DateTo),
		SortBy:       model.LessonSortBy(params.SortBy.Or(lessonsv1.LessonSortByStartAt)),
		SortOrder:    model.SortOrder(params.SortOrder.Or(lessonsv1.SortOrderAsc)),
		Limit:        params.Limit.Or(0),
		Offset:       params.Offset.Or(0),
	}
}

func ConvertBatchApplyRequest(req *lessonsv1.BatchApplyRequest) model.BatchApplyRequest {
	result := model.BatchApplyRequest{
		Mode:    model.BatchReplaceMode(req.Mode),
		Lessons: make([]model.BatchLessonUpsert, 0, len(req.Lessons)),
	}
	for _, item := range req.Lessons {
		result.Lessons = append(result.Lessons, model.BatchLessonUpsert{
			BranchID:     optNilUUIDToPtr(item.BranchID),
			StudentID:    item.StudentID.String(),
			InstructorID: item.InstructorID.String(),
			CarID:        item.CarID.String(),
			StartAt:      item.StartAt,
			EndAt:        item.EndAt,
			Comment:      optNilStringToPtr(item.Comment),
		})
	}
	return result
}

func ConvertBatchApplyResult(item model.BatchApplyResult) (lessonsv1.BatchApplyResponse, error) {
	lessons, err := ConvertLessons(item.Lessons)
	if err != nil {
		return lessonsv1.BatchApplyResponse{}, err
	}
	conflicts := make([]lessonsv1.BatchConflict, 0, len(item.Conflicts))
	for _, conflict := range item.Conflicts {
		conflicts = append(conflicts, lessonsv1.BatchConflict{
			Index:   conflict.Index,
			Code:    conflict.Code,
			Message: conflict.Message,
		})
	}
	return lessonsv1.BatchApplyResponse{
		Data: lessonsv1.BatchApplyResult{
			Lessons:    lessons,
			Conflicts:  conflicts,
			Applied:    item.Applied,
			Requested:  item.Requested,
			Mode:       lessonsv1.BatchApplyMode(item.Mode),
			Successful: item.Successful,
		},
	}, nil
}

func ConvertMeta(meta model.PageMeta) lessonsv1.Meta {
	return lessonsv1.Meta{
		Limit:  meta.Limit,
		Offset: meta.Offset,
		Count:  meta.Count,
	}
}

func optUUIDToPtr(v lessonsv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optNilUUIDToPtr(v lessonsv1.OptNilUUID) *string {
	if !v.IsSet() || v.IsNull() {
		return nil
	}
	value, ok := v.Get()
	if !ok {
		return nil
	}
	s := value.String()
	return &s
}

func optLessonStatusToPtr(v lessonsv1.OptLessonStatus) *model.LessonStatus {
	if value, ok := v.Get(); ok {
		status := model.LessonStatus(value)
		return &status
	}
	return nil
}

func optDateTimeToPtr(v lessonsv1.OptDateTime) *time.Time {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optNilStringToPtr(v lessonsv1.OptNilString) *string {
	if !v.IsSet() || v.IsNull() {
		return nil
	}
	value, ok := v.Get()
	if !ok {
		return nil
	}
	return &value
}

func optNilStringToPatch(v lessonsv1.OptNilString) (*string, bool) {
	if !v.IsSet() {
		return nil, false
	}
	if v.IsNull() {
		return nil, true
	}
	value, ok := v.Get()
	if !ok {
		return nil, true
	}
	return &value, true
}

func ptrStringToOptNilString(v *string) lessonsv1.OptNilString {
	if v == nil {
		return lessonsv1.OptNilString{}
	}
	var result lessonsv1.OptNilString
	result.SetTo(*v)
	return result
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}
