package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/lesson/internal/httpauth"
	"github.com/ossirdian/magistral/lesson/internal/planning"
	"github.com/ossirdian/magistral/lesson/internal/service"
	lessonsv1 "github.com/ossirdian/magistral/shared/pkg/lessons/v1"
)

var (
	errNilLessonService    = errors.New("lesson service is nil")
	errNilPlanningProducer = errors.New("planning producer is nil")
)

type api struct {
	lessonService    service.LessonService
	planningProducer *planning.Producer
}

func New(lessonService service.LessonService, planningProducer *planning.Producer) (*api, error) {
	if lessonService == nil {
		return nil, errNilLessonService
	}
	if planningProducer == nil {
		return nil, errNilPlanningProducer
	}
	return &api{lessonService: lessonService, planningProducer: planningProducer}, nil
}

func NewServer(lessonService service.LessonService, planningProducer *planning.Producer, opts ...lessonsv1.ServerOption) (*lessonsv1.Server, error) {
	handler, err := New(lessonService, planningProducer)
	if err != nil {
		return nil, err
	}
	return lessonsv1.NewServer(handler, opts...)
}

func (a *api) schoolID(ctx context.Context) (string, error) {
	schoolID, ok := httpauth.SchoolIDFromContext(ctx)
	if !ok || schoolID == "" {
		return "", service.ErrInvalidArgument
	}
	return schoolID, nil
}
