package api

import (
	"context"
	"errors"

	studentsv1 "github.com/ossirdian/magistral/shared/pkg/students/v1"
	"github.com/ossirdian/magistral/student/internal/converter"
	repo "github.com/ossirdian/magistral/student/internal/repository"
	svc "github.com/ossirdian/magistral/student/internal/service"
)

func (a *api) CreateStudent(ctx context.Context, req *studentsv1.StudentCreateRequest) (studentsv1.CreateStudentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.CreateStudentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
		return &res, nil
	}

	item, err := a.studentService.Create(ctx, schoolID, converter.ConvertStudentCreateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.CreateStudentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.CreateStudentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertStudent(item)
	if err != nil {
		return nil, err
	}

	return &studentsv1.StudentResponse{Data: data}, nil
}

func (a *api) DeleteStudent(ctx context.Context, params studentsv1.DeleteStudentParams) (studentsv1.DeleteStudentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.DeleteStudentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.studentService.Delete(ctx, schoolID, params.StudentId.String()); err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.DeleteStudentConflict(newErrorBody(ctx, "conflict", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.DeleteStudentNotFound(newErrorBody(ctx, "not_found", "student not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.DeleteStudentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &studentsv1.DeleteStudentNoContent{}, nil
}

func (a *api) GetStudent(ctx context.Context, params studentsv1.GetStudentParams) (studentsv1.GetStudentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.GetStudentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.studentService.Get(ctx, schoolID, params.StudentId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.GetStudentNotFound(newErrorBody(ctx, "not_found", "student not found"))
			return &res, nil
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.GetStudentForbidden(newErrorBody(ctx, "forbidden", "forbidden"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertStudent(item)
	if err != nil {
		return nil, err
	}

	return &studentsv1.StudentResponse{Data: data}, nil
}

func (a *api) ListStudents(ctx context.Context, params studentsv1.ListStudentsParams) (studentsv1.ListStudentsRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.ListStudentsUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.studentService.List(ctx, schoolID, converter.ConvertStudentFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := studentsv1.ListStudentsBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertStudents(items)
	if err != nil {
		return nil, err
	}

	return &studentsv1.StudentListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchStudent(ctx context.Context, req *studentsv1.StudentPatchRequest, params studentsv1.PatchStudentParams) (studentsv1.PatchStudentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := studentsv1.PatchStudentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.studentService.Update(ctx, schoolID, params.StudentId.String(), converter.ConvertStudentPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := studentsv1.PatchStudentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := studentsv1.PatchStudentNotFound(newErrorBody(ctx, "not_found", "student not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := studentsv1.PatchStudentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertStudent(item)
	if err != nil {
		return nil, err
	}

	return &studentsv1.StudentResponse{Data: data}, nil
}
