package service

import (
	"context"

	"github.com/ossirdian/magistral/student/internal/model"
)

type StudentService interface {
	Create(ctx context.Context, schoolID string, input model.StudentCreate) (model.Student, error)
	Get(ctx context.Context, schoolID, studentID string) (model.Student, error)
	List(ctx context.Context, schoolID string, filter model.StudentFilter) ([]model.Student, model.PageMeta, error)
	Update(ctx context.Context, schoolID, studentID string, patch model.StudentPatch) (model.Student, error)
	Delete(ctx context.Context, schoolID, studentID string) error
}

type StudyProgramService interface {
	Create(ctx context.Context, schoolID string, input model.StudyProgramCreate) (model.StudyProgram, error)
	Get(ctx context.Context, schoolID, programID string) (model.StudyProgram, error)
	List(ctx context.Context, schoolID string, filter model.StudyProgramFilter) ([]model.StudyProgram, model.PageMeta, error)
	Update(ctx context.Context, schoolID, programID string, patch model.StudyProgramPatch) (model.StudyProgram, error)
	Delete(ctx context.Context, schoolID, programID string) error
}

type GroupService interface {
	Create(ctx context.Context, schoolID string, input model.GroupCreate) (model.Group, error)
	Get(ctx context.Context, schoolID, groupID string) (model.Group, error)
	List(ctx context.Context, schoolID string, filter model.GroupFilter) ([]model.Group, model.PageMeta, error)
	Update(ctx context.Context, schoolID, groupID string, patch model.GroupPatch) (model.Group, error)
	Delete(ctx context.Context, schoolID, groupID string) error
}
