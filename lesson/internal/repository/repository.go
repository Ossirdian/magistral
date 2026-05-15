package repository

import (
	"context"
	"time"

	"github.com/ossirdian/magistral/lesson/internal/model"
)

type LessonRepository interface {
	GetByID(ctx context.Context, schoolID, lessonID string) (model.Lesson, error)
	List(ctx context.Context, schoolID string, filter model.LessonFilter) ([]model.Lesson, model.PageMeta, error)
	Create(ctx context.Context, schoolID string, input model.LessonCreate) (model.Lesson, error)
	Update(ctx context.Context, schoolID, lessonID string, patch model.LessonPatch) (model.Lesson, error)
	Delete(ctx context.Context, schoolID, lessonID string) error
	DeleteGeneratedInWindow(ctx context.Context, schoolID string, studentIDs []string, dateFrom, dateTo time.Time) error
}
