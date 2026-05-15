package api

import (
	"context"
	"errors"

	converter "github.com/ossirdian/magistral/resources/internal/converter"
	repo "github.com/ossirdian/magistral/resources/internal/repository"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
)

func (a *api) CreateInstructor(ctx context.Context, req *resourcesv1.InstructorCreateRequest) (resourcesv1.CreateInstructorRes, error) {
	item, err := a.instructorService.Create(ctx, a.schoolID(ctx), converter.ConvertInstructorCreateRequest(req))
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			res := resourcesv1.CreateInstructorConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertInstructor(item)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.InstructorResponse{Data: data}, nil
}

func (a *api) DeleteInstructor(ctx context.Context, params resourcesv1.DeleteInstructorParams) (resourcesv1.DeleteInstructorRes, error) {
	err := a.instructorService.Delete(ctx, a.schoolID(ctx), params.InstructorId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.DeleteInstructorNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.DeleteInstructorConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &resourcesv1.DeleteInstructorNoContent{}, nil
}

func (a *api) GetInstructor(ctx context.Context, params resourcesv1.GetInstructorParams) (resourcesv1.GetInstructorRes, error) {
	item, err := a.instructorService.Get(ctx, a.schoolID(ctx), params.InstructorId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := resourcesv1.GetInstructorNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertInstructor(item)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.InstructorResponse{Data: data}, nil
}

func (a *api) GetInstructorAvailability(ctx context.Context, params resourcesv1.GetInstructorAvailabilityParams) (resourcesv1.GetInstructorAvailabilityRes, error) {
	items, err := a.instructorService.GetAvailability(ctx, a.schoolID(ctx), params.InstructorId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := resourcesv1.GetInstructorAvailabilityNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertAvailabilityWindows(items)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.AvailabilityWindowListResponse{Data: data}, nil
}

func (a *api) ListInstructors(ctx context.Context, params resourcesv1.ListInstructorsParams) (resourcesv1.ListInstructorsRes, error) {
	items, meta, err := a.instructorService.List(ctx, a.schoolID(ctx), converter.ConvertInstructorFilter(params))
	if err != nil {
		return nil, err
	}

	data, err := converter.ConvertInstructors(items)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.InstructorListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchInstructor(ctx context.Context, req *resourcesv1.InstructorPatchRequest, params resourcesv1.PatchInstructorParams) (resourcesv1.PatchInstructorRes, error) {
	item, err := a.instructorService.Update(ctx, a.schoolID(ctx), params.InstructorId.String(), converter.ConvertInstructorPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.PatchInstructorNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.PatchInstructorConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertInstructor(item)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.InstructorResponse{Data: data}, nil
}

func (a *api) ReplaceInstructorAvailability(ctx context.Context, req *resourcesv1.AvailabilityReplaceRequest, params resourcesv1.ReplaceInstructorAvailabilityParams) (resourcesv1.ReplaceInstructorAvailabilityRes, error) {
	items, err := a.instructorService.ReplaceAvailability(ctx, a.schoolID(ctx), params.InstructorId.String(), converter.ConvertAvailabilityReplaceRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.ReplaceInstructorAvailabilityNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.ReplaceInstructorAvailabilityConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertAvailabilityWindows(items)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.AvailabilityWindowListResponse{Data: data}, nil
}
