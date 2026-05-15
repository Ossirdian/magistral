package converter

import (
	svcmodel "github.com/ossirdian/magistral/saas/internal/model"
	repomodel "github.com/ossirdian/magistral/saas/internal/repository/model"
)

func SchoolToModel(item repomodel.School) svcmodel.School {
	return svcmodel.School{
		ID:        item.ID,
		Name:      item.Name,
		Status:    svcmodel.SchoolStatus(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: item.DeletedAt,
	}
}

func SchoolsToModel(items []repomodel.School) []svcmodel.School {
	result := make([]svcmodel.School, 0, len(items))
	for _, item := range items {
		result = append(result, SchoolToModel(item))
	}
	return result
}

func SchoolCreateFromModel(input svcmodel.SchoolCreate) repomodel.SchoolCreate {
	return repomodel.SchoolCreate{
		Name:   input.Name,
		Status: schoolStatusPtrFromModel(input.Status),
	}
}

func SchoolPatchFromModel(input svcmodel.SchoolPatch) repomodel.SchoolPatch {
	return repomodel.SchoolPatch{
		Name:   input.Name,
		Status: schoolStatusPtrFromModel(input.Status),
	}
}

func SchoolFilterFromModel(input svcmodel.SchoolFilter) repomodel.SchoolFilter {
	return repomodel.SchoolFilter{
		Status: schoolStatusPtrFromModel(input.Status),
		Search: input.Search,
		Limit:  input.Limit,
		Offset: input.Offset,
	}
}
