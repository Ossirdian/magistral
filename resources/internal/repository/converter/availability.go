package converter

import (
	svcmodel "github.com/ossirdian/magistral/resources/internal/model"
	repomodel "github.com/ossirdian/magistral/resources/internal/repository/model"
)

func AvailabilityWindowToModel(item repomodel.AvailabilityWindow) svcmodel.AvailabilityWindow {
	return svcmodel.AvailabilityWindow{
		ID:           item.ID,
		ResourceType: svcmodel.ResourceType(item.ResourceType),
		ResourceID:   item.ResourceID,
		Type:         svcmodel.AvailabilityWindowType(item.Type),
		StartAt:      item.StartAt,
		EndAt:        item.EndAt,
		Comment:      item.Comment,
	}
}

func AvailabilityWindowsToModel(items []repomodel.AvailabilityWindow) []svcmodel.AvailabilityWindow {
	result := make([]svcmodel.AvailabilityWindow, 0, len(items))
	for _, item := range items {
		result = append(result, AvailabilityWindowToModel(item))
	}
	return result
}

func AvailabilityWindowInputFromModel(items []svcmodel.AvailabilityWindowInput) []repomodel.AvailabilityWindowInput {
	result := make([]repomodel.AvailabilityWindowInput, 0, len(items))
	for _, item := range items {
		result = append(result, repomodel.AvailabilityWindowInput{
			Type:    repomodel.AvailabilityWindowType(item.Type),
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
			Comment: item.Comment,
		})
	}
	return result
}
