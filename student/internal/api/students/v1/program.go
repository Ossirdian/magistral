package api

import (
	"context"
	"errors"

	studentsv1 "github.com/ossirdian/magistral/shared/pkg/students/v1"
	"github.com/ossirdian/magistral/student/internal/converter"
	repo "github.com/ossirdian/magistral/student/internal/repository"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (a *api) CreateStudyProgram(ctx context.Context, req *studentsv1.StudyProgramCreateRequest) (studentsv1.CreateStudyProgramRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.CreateStudyProgramBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
		return &res, nil
	}

	item, err := a.programService.Create(ctx, schoolID, converter.ConvertStudyProgramCreateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.CreateStudyProgramBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.CreateStudyProgramConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertStudyProgram(item)
	if err != nil {
		return nil, err
	}
	return &studentsv1.StudyProgramResponse{Data: data}, nil
}

func (a *api) DeleteStudyProgram(ctx context.Context, params studentsv1.DeleteStudyProgramParams) (studentsv1.DeleteStudyProgramRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.DeleteStudyProgramUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.programService.Delete(ctx, schoolID, params.ProgramId.String()); err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.DeleteStudyProgramConflict(newErrorBody(ctx, "conflict", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.DeleteStudyProgramNotFound(newErrorBody(ctx, "not_found", "study program not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.DeleteStudyProgramConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &studentsv1.DeleteStudyProgramNoContent{}, nil
}

func (a *api) GetStudyProgram(ctx context.Context, params studentsv1.GetStudyProgramParams) (studentsv1.GetStudyProgramRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.GetStudyProgramUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.programService.Get(ctx, schoolID, params.ProgramId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.GetStudyProgramNotFound(newErrorBody(ctx, "not_found", "study program not found"))
			return &res, nil
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.GetStudyProgramForbidden(newErrorBody(ctx, "forbidden", "forbidden"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertStudyProgram(item)
	if err != nil {
		return nil, err
	}
	return &studentsv1.StudyProgramResponse{Data: data}, nil
}

func (a *api) ListStudyPrograms(ctx context.Context, params studentsv1.ListStudyProgramsParams) (studentsv1.ListStudyProgramsRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.ListStudyProgramsUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.programService.List(ctx, schoolID, converter.ConvertStudyProgramFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := studentsv1.ListStudyProgramsBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertStudyPrograms(items)
	if err != nil {
		return nil, err
	}

	return &studentsv1.StudyProgramListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchStudyProgram(ctx context.Context, req *studentsv1.StudyProgramPatchRequest, params studentsv1.PatchStudyProgramParams) (studentsv1.PatchStudyProgramRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.PatchStudyProgramUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.programService.Update(ctx, schoolID, params.ProgramId.String(), converter.ConvertStudyProgramPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.PatchStudyProgramBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.PatchStudyProgramNotFound(newErrorBody(ctx, "not_found", "study program not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.PatchStudyProgramConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertStudyProgram(item)
	if err != nil {
		return nil, err
	}
	return &studentsv1.StudyProgramResponse{Data: data}, nil
}
