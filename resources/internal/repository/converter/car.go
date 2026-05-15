package converter

import (
	svcmodel "github.com/ossirdian/magistral/resources/internal/model"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func CarToModel(item repomodel.Car) svcmodel.Car {
	return svcmodel.Car{
		ID:                 item.ID,
		SchoolID:           item.SchoolID,
		BranchID:           item.BranchID,
		InstructorID:       item.InstructorID,
		PlateNumber:        item.PlateNumber,
		Brand:              item.Brand,
		Model:              item.Model,
		Color:              item.Color,
		Transmission:       svcmodel.TransmissionType(item.Transmission),
		Status:             svcmodel.CarStatus(item.Status),
		AvailabilityStatus: svcmodel.AvailabilityStatus(item.AvailabilityStatus),
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func CarsToModel(items []repomodel.Car) []svcmodel.Car {
	result := make([]svcmodel.Car, 0, len(items))
	for _, item := range items {
		result = append(result, CarToModel(item))
	}
	return result
}

func CarCreateFromModel(input svcmodel.CarCreate) repomodel.CarCreate {
	var status *repomodel.CarStatus
	if input.Status != nil {
		value := repomodel.CarStatus(*input.Status)
		status = &value
	}

	return repomodel.CarCreate{
		BranchID:     input.BranchID,
		InstructorID: input.InstructorID,
		PlateNumber:  input.PlateNumber,
		Brand:        input.Brand,
		Model:        input.Model,
		Color:        input.Color,
		Transmission: transmissionTypePtrFromModel(input.Transmission),
		Status:       status,
	}
}

func CarPatchFromModel(input svcmodel.CarPatch) repomodel.CarPatch {
	return repomodel.CarPatch{
		BranchID:     input.BranchID,
		InstructorID: stringFieldFromModel(input.InstructorID),
		PlateNumber:  input.PlateNumber,
		Brand:        stringFieldFromModel(input.Brand),
		Model:        stringFieldFromModel(input.Model),
		Color:        stringFieldFromModel(input.Color),
		Transmission: transmissionTypePtrFromModel(input.Transmission),
		Status:       carStatusPtrFromModel(input.Status),
	}
}

func transmissionTypePtrFromModel(input *svcmodel.TransmissionType) *repomodel.TransmissionType {
	if input == nil {
		return nil
	}
	value := repomodel.TransmissionType(*input)
	return &value
}

func CarFilterFromModel(input svcmodel.CarFilter) repomodel.CarFilter {
	return repomodel.CarFilter{
		BranchID:           input.BranchID,
		InstructorID:       input.InstructorID,
		Status:             carStatusPtrFromModel(input.Status),
		AvailabilityStatus: availabilityStatusPtrFromModel(input.AvailabilityStatus),
		Limit:              input.Limit,
		Offset:             input.Offset,
	}
}

func CarAssignmentFromModel(input svcmodel.CarAssignment) repomodel.CarAssignment {
	return repomodel.CarAssignment{
		InstructorID:  input.InstructorID,
		EffectiveFrom: input.EffectiveFrom,
		Comment:       input.Comment,
	}
}

func CarAssignmentResultToModel(item repomodel.CarAssignmentResult) svcmodel.CarAssignmentResult {
	return svcmodel.CarAssignmentResult{
		CarID:                item.CarID,
		InstructorID:         item.InstructorID,
		PreviousInstructorID: item.PreviousInstructorID,
		AssignedAt:           item.AssignedAt,
	}
}
