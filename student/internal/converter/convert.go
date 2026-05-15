package converter

import (
	"fmt"

	"github.com/google/uuid"
	studentsv1 "github.com/ossirdian/magistral/shared/pkg/students/v1"
	"github.com/ossirdian/magistral/student/internal/model"
)

func ConvertStudent(item model.Student) (studentsv1.Student, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return studentsv1.Student{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return studentsv1.Student{}, err
	}
	branchID, err := parseUUID(item.BranchID)
	if err != nil {
		return studentsv1.Student{}, err
	}

	result := studentsv1.Student{
		ID:                   id,
		SchoolID:             schoolID,
		BranchID:             branchID,
		FullName:             item.FullName,
		Phone:                item.Phone,
		Status:               studentsv1.StudentStatus(item.Status),
		PreferredWeekdays:    convertPreferredWeekdays(item.PreferredWeekdays),
		PreferredTimeWindows: convertPreferredTimeWindows(item.PreferredTimeWindows),
		ManualHoursCompleted: item.ManualHoursCompleted,
		LessonHoursCompleted: item.LessonHoursCompleted,
		TotalHoursCompleted:  item.TotalHoursCompleted,
		RemainingHours:       item.RemainingHours,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
	if item.GroupID != nil {
		groupID, err := parseUUID(*item.GroupID)
		if err != nil {
			return studentsv1.Student{}, err
		}
		result.GroupID.SetTo(groupID)
	} else {
		result.GroupID.SetToNull()
	}
	if item.ProgramID != nil {
		programID, err := parseUUID(*item.ProgramID)
		if err != nil {
			return studentsv1.Student{}, err
		}
		result.ProgramID.SetTo(programID)
	}
	return result, nil
}

func ConvertGroup(item model.Group) (studentsv1.Group, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return studentsv1.Group{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return studentsv1.Group{}, err
	}
	branchID, err := parseUUID(item.BranchID)
	if err != nil {
		return studentsv1.Group{}, err
	}

	return studentsv1.Group{
		ID:        id,
		SchoolID:  schoolID,
		BranchID:  branchID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

func ConvertStudyProgram(item model.StudyProgram) (studentsv1.StudyProgram, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return studentsv1.StudyProgram{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return studentsv1.StudyProgram{}, err
	}

	return studentsv1.StudyProgram{
		ID:           id,
		SchoolID:     schoolID,
		Name:         item.Name,
		Transmission: studentsv1.TransmissionType(item.Transmission),
		TotalHours:   item.TotalHours,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}, nil
}

func ConvertStudents(items []model.Student) ([]studentsv1.Student, error) {
	result := make([]studentsv1.Student, 0, len(items))
	for _, item := range items {
		converted, err := ConvertStudent(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertGroups(items []model.Group) ([]studentsv1.Group, error) {
	result := make([]studentsv1.Group, 0, len(items))
	for _, item := range items {
		converted, err := ConvertGroup(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertStudyPrograms(items []model.StudyProgram) ([]studentsv1.StudyProgram, error) {
	result := make([]studentsv1.StudyProgram, 0, len(items))
	for _, item := range items {
		converted, err := ConvertStudyProgram(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertStudentCreateRequest(req *studentsv1.StudentCreateRequest) model.StudentCreate {
	return model.StudentCreate{
		BranchID:             req.BranchID.String(),
		GroupID:              optNilUUIDToPtr(req.GroupID),
		ProgramID:            req.ProgramID.String(),
		FullName:             req.FullName,
		Phone:                req.Phone,
		Status:               optStatusToPtr(req.Status),
		PreferredWeekdays:    convertPreferredWeekdaysFromRequest(req.PreferredWeekdays),
		PreferredTimeWindows: convertPreferredTimeWindowsFromRequest(req.PreferredTimeWindows),
		ManualHoursCompleted: optFloat64ToPtr(req.ManualHoursCompleted),
	}
}

func ConvertGroupCreateRequest(req *studentsv1.GroupCreateRequest) model.GroupCreate {
	return model.GroupCreate{
		BranchID: req.BranchID.String(),
		Name:     req.Name,
	}
}

func ConvertStudentPatchRequest(req *studentsv1.StudentPatchRequest) model.StudentPatch {
	return model.StudentPatch{
		BranchID:             optUUIDToPtr(req.BranchID),
		GroupID:              optNilUUIDToPtr(req.GroupID),
		GroupIDSet:           req.GroupID.IsSet(),
		ProgramID:            optUUIDToPtr(req.ProgramID),
		FullName:             optStringToPtr(req.FullName),
		Phone:                optStringToPtr(req.Phone),
		Status:               optStatusToPtr(req.Status),
		PreferredWeekdays:    convertPreferredWeekdaysFromRequest(req.PreferredWeekdays),
		PreferredTimeWindows: convertPreferredTimeWindowsFromRequest(req.PreferredTimeWindows),
		ManualHoursCompleted: optFloat64ToPtr(req.ManualHoursCompleted),
	}
}

func ConvertStudyProgramCreateRequest(req *studentsv1.StudyProgramCreateRequest) model.StudyProgramCreate {
	return model.StudyProgramCreate{
		Name:         req.Name,
		Transmission: model.TransmissionType(req.Transmission),
		TotalHours:   req.TotalHours,
	}
}

func ConvertStudyProgramPatchRequest(req *studentsv1.StudyProgramPatchRequest) model.StudyProgramPatch {
	return model.StudyProgramPatch{
		Name:         optStringToPtr(req.Name),
		Transmission: optTransmissionToPtr(req.Transmission),
		TotalHours:   optFloat64ToPtr(req.TotalHours),
	}
}

func ConvertGroupPatchRequest(req *studentsv1.GroupPatchRequest) model.GroupPatch {
	return model.GroupPatch{
		BranchID: optUUIDToPtr(req.BranchID),
		Name:     optStringToPtr(req.Name),
	}
}

func ConvertStudentFilter(params studentsv1.ListStudentsParams) model.StudentFilter {
	return model.StudentFilter{
		BranchID: optUUIDToPtr(params.BranchID),
		GroupID:  optUUIDToPtr(params.GroupID),
		Status:   optStatusToPtr(params.Status),
		Search:   optStringToPtr(params.Search),
		Limit:    params.Limit.Or(0),
		Offset:   params.Offset.Or(0),
	}
}

func ConvertGroupFilter(params studentsv1.ListGroupsParams) model.GroupFilter {
	return model.GroupFilter{
		BranchID: optUUIDToPtr(params.BranchID),
		Search:   optStringToPtr(params.Search),
		Limit:    params.Limit.Or(0),
		Offset:   params.Offset.Or(0),
	}
}

func ConvertStudyProgramFilter(params studentsv1.ListStudyProgramsParams) model.StudyProgramFilter {
	return model.StudyProgramFilter{
		Search: params.Search.Or(""),
		Limit:  params.Limit.Or(0),
		Offset: params.Offset.Or(0),
	}
}

func ConvertMeta(meta model.PageMeta) studentsv1.Meta {
	return studentsv1.Meta{
		Limit:  meta.Limit,
		Offset: meta.Offset,
		Count:  meta.Count,
	}
}

func optUUIDToPtr(v studentsv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optNilUUIDToPtr(v studentsv1.OptNilUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optStringToPtr(v studentsv1.OptString) *string {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optStatusToPtr(v studentsv1.OptStudentStatus) *model.StudentStatus {
	if value, ok := v.Get(); ok {
		status := model.StudentStatus(value)
		return &status
	}
	return nil
}

func optTransmissionToPtr(v studentsv1.OptTransmissionType) *model.TransmissionType {
	if value, ok := v.Get(); ok {
		transmission := model.TransmissionType(value)
		return &transmission
	}
	return nil
}

func optFloat64ToPtr(v studentsv1.OptFloat64) *float64 {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}

func convertPreferredWeekdays(values []model.StudentPreferredWeekday) []studentsv1.StudentPreferredWeekday {
	result := make([]studentsv1.StudentPreferredWeekday, 0, len(values))
	for _, value := range values {
		result = append(result, studentsv1.StudentPreferredWeekday(value))
	}
	return result
}

func convertPreferredTimeWindows(values []model.StudentPreferredTimeWindow) []studentsv1.StudentPreferredTimeWindow {
	result := make([]studentsv1.StudentPreferredTimeWindow, 0, len(values))
	for _, value := range values {
		result = append(result, studentsv1.StudentPreferredTimeWindow(value))
	}
	return result
}

func convertPreferredWeekdaysFromRequest(values []studentsv1.StudentPreferredWeekday) []model.StudentPreferredWeekday {
	result := make([]model.StudentPreferredWeekday, 0, len(values))
	for _, value := range values {
		result = append(result, model.StudentPreferredWeekday(value))
	}
	return result
}

func convertPreferredTimeWindowsFromRequest(values []studentsv1.StudentPreferredTimeWindow) []model.StudentPreferredTimeWindow {
	result := make([]model.StudentPreferredTimeWindow, 0, len(values))
	for _, value := range values {
		result = append(result, model.StudentPreferredTimeWindow(value))
	}
	return result
}
