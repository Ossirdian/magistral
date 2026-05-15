package converter

import (
	svcmodel "github.com/ossirdian/magistral/saas/internal/model"
	repomodel "github.com/ossirdian/magistral/saas/internal/repository/model"
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

func schoolStatusPtrFromModel(input *svcmodel.SchoolStatus) *repomodel.SchoolStatus {
	if input == nil {
		return nil
	}
	value := repomodel.SchoolStatus(*input)
	return &value
}

func branchStatusPtrFromModel(input *svcmodel.BranchStatus) *repomodel.BranchStatus {
	if input == nil {
		return nil
	}
	value := repomodel.BranchStatus(*input)
	return &value
}
