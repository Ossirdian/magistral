package api

import (
	"context"
	"errors"

	studentsv1 "github.com/ossirdian/magistral/shared/pkg/students/v1"
	"github.com/ossirdian/magistral/student/internal/converter"
	repo "github.com/ossirdian/magistral/student/internal/repository"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (a *api) CreateGroup(ctx context.Context, req *studentsv1.GroupCreateRequest) (studentsv1.CreateGroupRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.CreateGroupBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
		return &res, nil
	}

	item, err := a.groupService.Create(ctx, schoolID, converter.ConvertGroupCreateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.CreateGroupBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.CreateGroupConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertGroup(item)
	if err != nil {
		return nil, err
	}

	return &studentsv1.GroupResponse{Data: data}, nil
}

func (a *api) DeleteGroup(ctx context.Context, params studentsv1.DeleteGroupParams) (studentsv1.DeleteGroupRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.DeleteGroupUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.groupService.Delete(ctx, schoolID, params.GroupId.String()); err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.DeleteGroupConflict(newErrorBody(ctx, "conflict", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.DeleteGroupNotFound(newErrorBody(ctx, "not_found", "group not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.DeleteGroupConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &studentsv1.DeleteGroupNoContent{}, nil
}

func (a *api) GetGroup(ctx context.Context, params studentsv1.GetGroupParams) (studentsv1.GetGroupRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.GetGroupUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.groupService.Get(ctx, schoolID, params.GroupId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.GetGroupNotFound(newErrorBody(ctx, "not_found", "group not found"))
			return &res, nil
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.GetGroupForbidden(newErrorBody(ctx, "forbidden", "forbidden"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertGroup(item)
	if err != nil {
		return nil, err
	}

	return &studentsv1.GroupResponse{Data: data}, nil
}

func (a *api) ListGroups(ctx context.Context, params studentsv1.ListGroupsParams) (studentsv1.ListGroupsRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.ListGroupsUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.groupService.List(ctx, schoolID, converter.ConvertGroupFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := studentsv1.ListGroupsBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertGroups(items)
	if err != nil {
		return nil, err
	}

	return &studentsv1.GroupListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchGroup(ctx context.Context, req *studentsv1.GroupPatchRequest, params studentsv1.PatchGroupParams) (studentsv1.PatchGroupRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.PatchGroupUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.groupService.Update(ctx, schoolID, params.GroupId.String(), converter.ConvertGroupPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.PatchGroupBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.PatchGroupNotFound(newErrorBody(ctx, "not_found", "group not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.PatchGroupConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertGroup(item)
	if err != nil {
		return nil, err
	}

	return &studentsv1.GroupResponse{Data: data}, nil
}
