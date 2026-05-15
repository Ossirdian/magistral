package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/lesson/internal/converter"
	repo "github.com/ossirdian/magistral/lesson/internal/repository"
	svc "github.com/ossirdian/magistral/lesson/internal/service"
	lessonsv1 "github.com/ossirdian/magistral/shared/pkg/lessons/v1"
)

func (a *api) CreateLesson(ctx context.Context, req *lessonsv1.LessonCreateRequest) (lessonsv1.CreateLessonRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := lessonsv1.CreateLessonBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
		return &res, nil
	}

	item, err := a.lessonService.Create(ctx, schoolID, converter.ConvertLessonCreateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := lessonsv1.CreateLessonBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := lessonsv1.CreateLessonConflict(newErrorBody(ctx, "conflict", "lesson conflicts with schedule"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertLesson(item)
	if err != nil {
		return nil, err
	}

	return &lessonsv1.LessonResponse{Data: data}, nil
}

func (a *api) DeleteLesson(ctx context.Context, params lessonsv1.DeleteLessonParams) (lessonsv1.DeleteLessonRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := lessonsv1.DeleteLessonUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.lessonService.Delete(ctx, schoolID, params.LessonId.String()); err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := lessonsv1.DeleteLessonConflict(newErrorBody(ctx, "conflict", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := lessonsv1.DeleteLessonNotFound(newErrorBody(ctx, "not_found", "lesson not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := lessonsv1.DeleteLessonConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	return &lessonsv1.DeleteLessonNoContent{}, nil
}

func (a *api) GetLesson(ctx context.Context, params lessonsv1.GetLessonParams) (lessonsv1.GetLessonRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := lessonsv1.GetLessonUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.lessonService.Get(ctx, schoolID, params.LessonId.String())
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			res := lessonsv1.GetLessonNotFound(newErrorBody(ctx, "not_found", "lesson not found"))
			return &res, nil
		case errors.Is(err, svc.ErrInvalidArgument):
			res := lessonsv1.GetLessonForbidden(newErrorBody(ctx, "forbidden", "forbidden"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertLesson(item)
	if err != nil {
		return nil, err
	}

	return &lessonsv1.LessonResponse{Data: data}, nil
}

func (a *api) ListLessons(ctx context.Context, params lessonsv1.ListLessonsParams) (lessonsv1.ListLessonsRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := lessonsv1.ListLessonsUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.lessonService.List(ctx, schoolID, converter.ConvertLessonFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := lessonsv1.ListLessonsBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertLessons(items)
	if err != nil {
		return nil, err
	}

	return &lessonsv1.LessonListResponse{
		Data: data,
		Meta: converter.ConvertMeta(meta),
	}, nil
}

func (a *api) PatchLesson(ctx context.Context, req *lessonsv1.LessonPatchRequest, params lessonsv1.PatchLessonParams) (lessonsv1.PatchLessonRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := lessonsv1.PatchLessonUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.lessonService.Update(ctx, schoolID, params.LessonId.String(), converter.ConvertLessonPatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := lessonsv1.PatchLessonBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := lessonsv1.PatchLessonNotFound(newErrorBody(ctx, "not_found", "lesson not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := lessonsv1.PatchLessonConflict(newErrorBody(ctx, "conflict", "lesson conflicts with schedule"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertLesson(item)
	if err != nil {
		return nil, err
	}

	return &lessonsv1.LessonResponse{Data: data}, nil
}

func (a *api) ApplyLessonsBatch(ctx context.Context, req *lessonsv1.BatchApplyRequest) (lessonsv1.ApplyLessonsBatchRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := lessonsv1.ApplyLessonsBatchUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	result, err := a.lessonService.ApplyBatch(ctx, schoolID, converter.ConvertBatchApplyRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := lessonsv1.ApplyLessonsBatchBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertBatchApplyResult(result)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
