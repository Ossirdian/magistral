package converter

import (
	svcmodel "github.com/ossirdian/magistral/resources/internal/model"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func InstructorToModel(item repomodel.Instructor) svcmodel.Instructor {
	return svcmodel.Instructor{
		ID:                 item.ID,
		SchoolID:           item.SchoolID,
		BranchID:           item.BranchID,
		FullName:           item.FullName,
		Phone:              item.Phone,
		Email:              item.Email,
		Status:             svcmodel.InstructorStatus(item.Status),
		AvailabilityStatus: svcmodel.AvailabilityStatus(item.AvailabilityStatus),
		CurrentCarID:       item.CurrentCarID,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func InstructorsToModel(items []repomodel.Instructor) []svcmodel.Instructor {
	result := make([]svcmodel.Instructor, 0, len(items))
	for _, item := range items {
		result = append(result, InstructorToModel(item))
	}
	return result
}

func InstructorCreateFromModel(input svcmodel.InstructorCreate) repomodel.InstructorCreate {
	var status *repomodel.InstructorStatus
	if input.Status != nil {
		value := repomodel.InstructorStatus(*input.Status)
		status = &value
	}

	return repomodel.InstructorCreate{
		BranchID: input.BranchID,
		FullName: input.FullName,
		Phone:    input.Phone,
		Email:    input.Email,
		Status:   status,
	}
}

func InstructorPatchFromModel(input svcmodel.InstructorPatch) repomodel.InstructorPatch {
	return repomodel.InstructorPatch{
		BranchID: input.BranchID,
		FullName: input.FullName,
		Phone:    stringFieldFromModel(input.Phone),
		Email:    stringFieldFromModel(input.Email),
		Status:   instructorStatusPtrFromModel(input.Status),
	}
}

func InstructorFilterFromModel(input svcmodel.InstructorFilter) repomodel.InstructorFilter {
	return repomodel.InstructorFilter{
		BranchID:           input.BranchID,
		Status:             instructorStatusPtrFromModel(input.Status),
		AvailabilityStatus: availabilityStatusPtrFromModel(input.AvailabilityStatus),
		Limit:              input.Limit,
		Offset:             input.Offset,
	}
}
