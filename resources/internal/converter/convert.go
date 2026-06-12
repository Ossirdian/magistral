package converter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/resources/internal/model"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
)

func ConvertCar(item model.Car) (resourcesv1.Car, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return resourcesv1.Car{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return resourcesv1.Car{}, err
	}
	branchID, err := parseUUID(item.BranchID)
	if err != nil {
		return resourcesv1.Car{}, err
	}

	return resourcesv1.Car{
		ID:                 id,
		SchoolID:           schoolID,
		BranchID:           branchID,
		InstructorID:       mustOptNilUUIDFromPtr(item.InstructorID),
		PlateNumber:        item.PlateNumber,
		Brand:              optNilStringFromPtr(item.Brand),
		Model:              optNilStringFromPtr(item.Model),
		Color:              optNilStringFromPtr(item.Color),
		Transmission:       resourcesv1.TransmissionType(item.Transmission),
		Status:             resourcesv1.CarStatus(item.Status),
		AvailabilityStatus: resourcesv1.AvailabilityStatus(item.AvailabilityStatus),
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}, nil
}

func ConvertCars(items []model.Car) ([]resourcesv1.Car, error) {
	result := make([]resourcesv1.Car, 0, len(items))
	for _, item := range items {
		converted, err := ConvertCar(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertInstructor(item model.Instructor) (resourcesv1.Instructor, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return resourcesv1.Instructor{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return resourcesv1.Instructor{}, err
	}
	branchID, err := parseUUID(item.BranchID)
	if err != nil {
		return resourcesv1.Instructor{}, err
	}

	return resourcesv1.Instructor{
		ID:                 id,
		SchoolID:           schoolID,
		BranchID:           branchID,
		FullName:           item.FullName,
		Phone:              optNilStringFromPtr(item.Phone),
		Email:              optNilStringFromPtr(item.Email),
		Status:             resourcesv1.InstructorStatus(item.Status),
		AvailabilityStatus: resourcesv1.AvailabilityStatus(item.AvailabilityStatus),
		CurrentCarID:       mustOptNilUUIDFromPtr(item.CurrentCarID),
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}, nil
}

func ConvertInstructors(items []model.Instructor) ([]resourcesv1.Instructor, error) {
	result := make([]resourcesv1.Instructor, 0, len(items))
	for _, item := range items {
		converted, err := ConvertInstructor(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertAvailabilityWindow(item model.AvailabilityWindow) (resourcesv1.AvailabilityWindow, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return resourcesv1.AvailabilityWindow{}, err
	}
	resourceID, err := parseUUID(item.ResourceID)
	if err != nil {
		return resourcesv1.AvailabilityWindow{}, err
	}

	return resourcesv1.AvailabilityWindow{
		ID:           id,
		ResourceType: resourcesv1.ResourceType(item.ResourceType),
		ResourceID:   resourceID,
		Type:         resourcesv1.AvailabilityWindowType(item.Type),
		StartAt:      item.StartAt,
		EndAt:        item.EndAt,
		Comment:      optNilStringFromPtr(item.Comment),
	}, nil
}

func ConvertAvailabilityWindows(items []model.AvailabilityWindow) ([]resourcesv1.AvailabilityWindow, error) {
	result := make([]resourcesv1.AvailabilityWindow, 0, len(items))
	for _, item := range items {
		converted, err := ConvertAvailabilityWindow(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertCarAssignmentResult(item model.CarAssignmentResult) (resourcesv1.CarAssignmentResult, error) {
	carID, err := parseUUID(item.CarID)
	if err != nil {
		return resourcesv1.CarAssignmentResult{}, err
	}
	instructorID, err := parseUUID(item.InstructorID)
	if err != nil {
		return resourcesv1.CarAssignmentResult{}, err
	}

	return resourcesv1.CarAssignmentResult{
		CarID:                carID,
		InstructorID:         instructorID,
		PreviousInstructorID: mustOptNilUUIDFromPtr(item.PreviousInstructorID),
		AssignedAt:           item.AssignedAt,
	}, nil
}

func ConvertCarCreateRequest(req *resourcesv1.CarCreateRequest) model.CarCreate {
	return model.CarCreate{
		BranchID:     req.BranchID.String(),
		InstructorID: nil,
		PlateNumber:  req.PlateNumber,
		Brand:        optNilStringToPtr(req.Brand),
		Model:        optNilStringToPtr(req.Model),
		Color:        optNilStringToPtr(req.Color),
		Transmission: optTransmissionTypeToPtr(req.Transmission),
		Status:       optCarStatusToPtr(req.Status),
	}
}

func ConvertCarPatchRequest(req *resourcesv1.CarPatchRequest) model.CarPatch {
	return model.CarPatch{
		BranchID:     optUUIDToPtr(req.BranchID),
		PlateNumber:  optStringToPtr(req.PlateNumber),
		Brand:        stringFieldFromOptNilString(req.Brand),
		Model:        stringFieldFromOptNilString(req.Model),
		Color:        stringFieldFromOptNilString(req.Color),
		Transmission: optTransmissionTypeToPtr(req.Transmission),
		Status:       optCarStatusToPtr(req.Status),
	}
}

func ConvertInstructorCreateRequest(req *resourcesv1.InstructorCreateRequest) model.InstructorCreate {
	return model.InstructorCreate{
		BranchID: req.BranchID.String(),
		FullName: req.FullName,
		Phone:    optNilStringToPtr(req.Phone),
		Email:    optNilStringToPtr(req.Email),
		Status:   optInstructorStatusToPtr(req.Status),
	}
}

func ConvertInstructorPatchRequest(req *resourcesv1.InstructorPatchRequest) model.InstructorPatch {
	return model.InstructorPatch{
		BranchID: optUUIDToPtr(req.BranchID),
		FullName: optStringToPtr(req.FullName),
		Phone:    stringFieldFromOptNilString(req.Phone),
		Email:    stringFieldFromOptNilString(req.Email),
		Status:   optInstructorStatusToPtr(req.Status),
	}
}

func ConvertAvailabilityReplaceRequest(req *resourcesv1.AvailabilityReplaceRequest) []model.AvailabilityWindowInput {
	result := make([]model.AvailabilityWindowInput, 0, len(req.Windows))
	for _, window := range req.Windows {
		result = append(result, model.AvailabilityWindowInput{
			Type:    model.AvailabilityWindowType(window.Type),
			StartAt: window.StartAt,
			EndAt:   window.EndAt,
			Comment: optNilStringToPtr(window.Comment),
		})
	}
	return result
}

func ConvertCarAssignmentRequest(req *resourcesv1.CarAssignmentRequest) model.CarAssignment {
	return model.CarAssignment{
		InstructorID:  req.InstructorID.String(),
		EffectiveFrom: optNilDateTimeToPtr(req.EffectiveFrom),
		Comment:       optNilStringToPtr(req.Comment),
	}
}

func ConvertCarFilter(params resourcesv1.ListCarsParams) model.CarFilter {
	return model.CarFilter{
		BranchID:           optUUIDToPtr(params.BranchID),
		InstructorID:       optUUIDToPtr(params.InstructorID),
		Status:             optCarStatusToPtr(params.Status),
		AvailabilityStatus: optAvailabilityStatusToPtr(params.AvailabilityStatus),
		Limit:              params.Limit.Or(0),
		Offset:             params.Offset.Or(0),
	}
}

func ConvertInstructorFilter(params resourcesv1.ListInstructorsParams) model.InstructorFilter {
	return model.InstructorFilter{
		BranchID:           optUUIDToPtr(params.BranchID),
		Status:             optInstructorStatusToPtr(params.Status),
		AvailabilityStatus: optAvailabilityStatusToPtr(params.AvailabilityStatus),
		Limit:              params.Limit.Or(0),
		Offset:             params.Offset.Or(0),
	}
}

func ConvertMeta(meta model.PageMeta) resourcesv1.Meta {
	return resourcesv1.Meta{
		Limit:  meta.Limit,
		Offset: meta.Offset,
		Count:  meta.Count,
	}
}

func optNilStringToPtr(v resourcesv1.OptNilString) *string {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optNilDateTimeToPtr(v resourcesv1.OptNilDateTime) *time.Time {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optUUIDToPtr(v resourcesv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optStringToPtr(v resourcesv1.OptString) *string {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optNilUUIDToPtr(v resourcesv1.OptNilUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optCarStatusToPtr(v resourcesv1.OptCarStatus) *model.CarStatus {
	if value, ok := v.Get(); ok {
		status := model.CarStatus(value)
		return &status
	}
	return nil
}

func optTransmissionTypeToPtr(v resourcesv1.OptTransmissionType) *model.TransmissionType {
	if value, ok := v.Get(); ok {
		transmission := model.TransmissionType(value)
		return &transmission
	}
	return nil
}

func optInstructorStatusToPtr(v resourcesv1.OptInstructorStatus) *model.InstructorStatus {
	if value, ok := v.Get(); ok {
		status := model.InstructorStatus(value)
		return &status
	}
	return nil
}

func optAvailabilityStatusToPtr(v resourcesv1.OptAvailabilityStatus) *model.AvailabilityStatus {
	if value, ok := v.Get(); ok {
		status := model.AvailabilityStatus(value)
		return &status
	}
	return nil
}

func stringFieldFromOptNilString(v resourcesv1.OptNilString) model.Field[string] {
	field := model.Field[string]{Set: v.IsSet()}
	if value, ok := v.Get(); ok {
		field.Value = &value
	}
	return field
}

func stringFieldFromOptNilUUID(v resourcesv1.OptNilUUID) model.Field[string] {
	field := model.Field[string]{Set: v.IsSet()}
	if value, ok := v.Get(); ok {
		s := value.String()
		field.Value = &s
	}
	return field
}

func optNilStringFromPtr(v *string) resourcesv1.OptNilString {
	if v == nil {
		var result resourcesv1.OptNilString
		result.SetToNull()
		return result
	}
	return resourcesv1.NewOptNilString(*v)
}

func mustOptNilUUIDFromPtr(v *string) resourcesv1.OptNilUUID {
	var result resourcesv1.OptNilUUID
	if v == nil {
		result.SetToNull()
		return result
	}
	parsed, err := uuid.Parse(*v)
	if err != nil {
		panic(fmt.Sprintf("invalid uuid in model: %q", *v))
	}
	result.SetTo(parsed)
	return result
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}
