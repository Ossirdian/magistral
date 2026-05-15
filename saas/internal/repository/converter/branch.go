package converter

import (
	svcmodel "github.com/ossirdian/magistral/saas/internal/model"
	repomodel "github.com/ossirdian/magistral/saas/internal/repository/model"
)

func BranchToModel(item repomodel.Branch) svcmodel.Branch {
	return svcmodel.Branch{
		ID:        item.ID,
		SchoolID:  item.SchoolID,
		Name:      item.Name,
		Address:   item.Address,
		Status:    svcmodel.BranchStatus(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		DeletedAt: item.DeletedAt,
	}
}

func BranchesToModel(items []repomodel.Branch) []svcmodel.Branch {
	result := make([]svcmodel.Branch, 0, len(items))
	for _, item := range items {
		result = append(result, BranchToModel(item))
	}
	return result
}

func BranchCreateFromModel(input svcmodel.BranchCreate) repomodel.BranchCreate {
	return repomodel.BranchCreate{
		SchoolID: input.SchoolID,
		Name:     input.Name,
		Address:  input.Address,
		Status:   branchStatusPtrFromModel(input.Status),
	}
}

func BranchPatchFromModel(input svcmodel.BranchPatch) repomodel.BranchPatch {
	return repomodel.BranchPatch{
		SchoolID: input.SchoolID,
		Name:     input.Name,
		Address:  stringFieldFromModel(input.Address),
		Status:   branchStatusPtrFromModel(input.Status),
	}
}

func BranchFilterFromModel(input svcmodel.BranchFilter) repomodel.BranchFilter {
	return repomodel.BranchFilter{
		SchoolID: input.SchoolID,
		Status:   branchStatusPtrFromModel(input.Status),
		Limit:    input.Limit,
		Offset:   input.Offset,
	}
}
