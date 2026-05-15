package service

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/lesson/internal/model"
)

var ErrInvalidArgument = errors.New("invalid argument")

type LessonService interface {
	Create(ctx context.Context, schoolID string, input model.LessonCreate) (model.Lesson, error)
	Get(ctx context.Context, schoolID, lessonID string) (model.Lesson, error)
	List(ctx context.Context, schoolID string, filter model.LessonFilter) ([]model.Lesson, model.PageMeta, error)
	Update(ctx context.Context, schoolID, lessonID string, patch model.LessonPatch) (model.Lesson, error)
	Delete(ctx context.Context, schoolID, lessonID string) error
	ApplyBatch(ctx context.Context, schoolID string, input model.BatchApplyRequest) (model.BatchApplyResult, error)
}
