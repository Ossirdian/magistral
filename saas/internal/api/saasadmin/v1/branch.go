package api

import (
	"context"
	"errors"

	converter "github.com/ossirdian/magistral/saas/internal/converter"
	repo "github.com/ossirdian/magistral/saas/internal/repository"
	saasadminv1 "github.com/ossirdian/magistral/shared/pkg/saasadmin/v1"
)

func (a *api) CreateBranch(ctx context.Context, req *saasadminv1.BranchCreateRequest) (saasadminv1.CreateBranchRes, error) {
	item, err := a.branchService.Create(ctx, converter.ConvertBranchCreateRequest(req))
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			res := saasadminv1.CreateBranchConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertBranch(item)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.BranchResponse{Data: data}, nil
}

func (a *api) DeleteBranch(ctx context.Context, params saasadminv1.DeleteBranchParams) (saasadminv1.DeleteBranchRes, error) {
	err := a.branchService.Delete(ctx, params.BranchId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := saasadminv1.DeleteBranchNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := saasadminv1.DeleteBranchConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &saasadminv1.DeleteBranchNoContent{}, nil
}

func (a *api) GetBranch(ctx context.Context, params saasadminv1.GetBranchParams) (saasadminv1.GetBranchRes, error) {
	item, err := a.branchService.Get(ctx, params.BranchId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := saasadminv1.GetBranchNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertBranch(item)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.BranchResponse{Data: data}, nil
}

func (a *api) ListBranches(ctx context.Context, params saasadminv1.ListBranchesParams) (saasadminv1.ListBranchesRes, error) {
	items, meta, err := a.branchService.List(ctx, converter.ConvertBranchFilter(params))
	if err != nil {
		return nil, err
	}

	data, err := converter.ConvertBranches(items)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.BranchListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchBranch(ctx context.Context, req *saasadminv1.BranchPatchRequest, params saasadminv1.PatchBranchParams) (saasadminv1.PatchBranchRes, error) {
	item, err := a.branchService.Update(ctx, params.BranchId.String(), converter.ConvertBranchPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := saasadminv1.PatchBranchNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := saasadminv1.PatchBranchConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertBranch(item)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.BranchResponse{Data: data}, nil
}
