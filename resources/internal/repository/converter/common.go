package converter

import (
	svcmodel "github.com/ossirdian/magistral/resources/internal/model"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func PageMetaToModel(item repomodel.PageMeta) svcmodel.PageMeta {
	return svcmodel.PageMeta{
		Limit:  item.Limit,
		Offset: item.Offset,
		Count:  item.Count,
	}
}

func stringFieldFromModel(input svcmodel.Field[string]) repomodel.Field[string] {
	return repomodel.Field[string]{
		Set:   input.Set,
		Value: input.Value,
	}
}

func carStatusPtrFromModel(input *svcmodel.CarStatus) *repomodel.CarStatus {
	if input == nil {
		return nil
	}
	value := repomodel.CarStatus(*input)
	return &value
}

func instructorStatusPtrFromModel(input *svcmodel.InstructorStatus) *repomodel.InstructorStatus {
	if input == nil {
		return nil
	}
	value := repomodel.InstructorStatus(*input)
	return &value
}

func availabilityStatusPtrFromModel(input *svcmodel.AvailabilityStatus) *repomodel.AvailabilityStatus {
	if input == nil {
		return nil
	}
	value := repomodel.AvailabilityStatus(*input)
	return &value
}
