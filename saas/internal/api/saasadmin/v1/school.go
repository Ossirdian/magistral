package api

import (
	"context"
	"errors"

	converter "github.com/ossirdian/magistral/saas/internal/converter"
	repo "github.com/ossirdian/magistral/saas/internal/repository"
	saasadminv1 "github.com/ossirdian/magistral/shared/pkg/saasadmin/v1"
)

func (a *api) CreateSchool(ctx context.Context, req *saasadminv1.SchoolCreateRequest) (saasadminv1.CreateSchoolRes, error) {
	item, err := a.schoolService.Create(ctx, converter.ConvertSchoolCreateRequest(req))
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			res := saasadminv1.CreateSchoolConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertSchool(item)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.SchoolResponse{Data: data}, nil
}

func (a *api) DeleteSchool(ctx context.Context, params saasadminv1.DeleteSchoolParams) (saasadminv1.DeleteSchoolRes, error) {
	err := a.schoolService.Delete(ctx, params.SchoolId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := saasadminv1.DeleteSchoolNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := saasadminv1.DeleteSchoolConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &saasadminv1.DeleteSchoolNoContent{}, nil
}

func (a *api) GetSchool(ctx context.Context, params saasadminv1.GetSchoolParams) (saasadminv1.GetSchoolRes, error) {
	item, err := a.schoolService.Get(ctx, params.SchoolId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := saasadminv1.GetSchoolNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertSchool(item)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.SchoolResponse{Data: data}, nil
}

func (a *api) ListSchools(ctx context.Context, params saasadminv1.ListSchoolsParams) (saasadminv1.ListSchoolsRes, error) {
	items, meta, err := a.schoolService.List(ctx, converter.ConvertSchoolFilter(params))
	if err != nil {
		return nil, err
	}

	data, err := converter.ConvertSchools(items)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.SchoolListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchSchool(ctx context.Context, req *saasadminv1.SchoolPatchRequest, params saasadminv1.PatchSchoolParams) (saasadminv1.PatchSchoolRes, error) {
	item, err := a.schoolService.Update(ctx, params.SchoolId.String(), converter.ConvertSchoolPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := saasadminv1.PatchSchoolNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := saasadminv1.PatchSchoolConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertSchool(item)
	if err != nil {
		return nil, err
	}

	return &saasadminv1.SchoolResponse{Data: data}, nil
}
