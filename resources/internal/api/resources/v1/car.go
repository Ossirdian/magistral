package api

import (
	"context"
	"errors"

	converter "github.com/ossirdian/magistral/resources/internal/converter"
	repo "github.com/ossirdian/magistral/resources/internal/repository"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
)

func (a *api) AssignCar(ctx context.Context, req *resourcesv1.CarAssignmentRequest, params resourcesv1.AssignCarParams) (resourcesv1.AssignCarRes, error) {
	result, err := a.carService.AssignInstructor(ctx, a.schoolID(ctx), params.CarId.String(), converter.ConvertCarAssignmentRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.AssignCarNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.AssignCarConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertCarAssignmentResult(result)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.CarAssignmentResponse{Data: data}, nil
}

func (a *api) CreateCar(ctx context.Context, req *resourcesv1.CarCreateRequest) (resourcesv1.CreateCarRes, error) {
	item, err := a.carService.Create(ctx, a.schoolID(ctx), converter.ConvertCarCreateRequest(req))
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			res := resourcesv1.CreateCarConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertCar(item)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.CarResponse{Data: data}, nil
}

func (a *api) DeleteCar(ctx context.Context, params resourcesv1.DeleteCarParams) (resourcesv1.DeleteCarRes, error) {
	err := a.carService.Delete(ctx, a.schoolID(ctx), params.CarId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.DeleteCarNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.DeleteCarConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &resourcesv1.DeleteCarNoContent{}, nil
}

func (a *api) GetCar(ctx context.Context, params resourcesv1.GetCarParams) (resourcesv1.GetCarRes, error) {
	item, err := a.carService.Get(ctx, a.schoolID(ctx), params.CarId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := resourcesv1.GetCarNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertCar(item)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.CarResponse{Data: data}, nil
}

func (a *api) GetCarAvailability(ctx context.Context, params resourcesv1.GetCarAvailabilityParams) (resourcesv1.GetCarAvailabilityRes, error) {
	items, err := a.carService.GetAvailability(ctx, a.schoolID(ctx), params.CarId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := resourcesv1.GetCarAvailabilityNotFound(newErrorBody(ctx, "not_found", "resource not found"))
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

func (a *api) ListCars(ctx context.Context, params resourcesv1.ListCarsParams) (resourcesv1.ListCarsRes, error) {
	items, meta, err := a.carService.List(ctx, a.schoolID(ctx), converter.ConvertCarFilter(params))
	if err != nil {
		return nil, err
	}

	data, err := converter.ConvertCars(items)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.CarListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchCar(ctx context.Context, req *resourcesv1.CarPatchRequest, params resourcesv1.PatchCarParams) (resourcesv1.PatchCarRes, error) {
	item, err := a.carService.Update(ctx, a.schoolID(ctx), params.CarId.String(), converter.ConvertCarPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.PatchCarNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.PatchCarConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertCar(item)
	if err != nil {
		return nil, err
	}

	return &resourcesv1.CarResponse{Data: data}, nil
}

func (a *api) ReplaceCarAvailability(ctx context.Context, req *resourcesv1.AvailabilityReplaceRequest, params resourcesv1.ReplaceCarAvailabilityParams) (resourcesv1.ReplaceCarAvailabilityRes, error) {
	items, err := a.carService.ReplaceAvailability(ctx, a.schoolID(ctx), params.CarId.String(), converter.ConvertAvailabilityReplaceRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := resourcesv1.ReplaceCarAvailabilityNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := resourcesv1.ReplaceCarAvailabilityConflict(newErrorBody(ctx, "conflict", "resource conflict"))
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
