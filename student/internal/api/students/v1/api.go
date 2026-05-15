package api

import (
	"context"
	"errors"

	studentsv1 "github.com/ossirdian/magistral/shared/pkg/students/v1"
	"github.com/ossirdian/magistral/student/internal/httpauth"
	"github.com/ossirdian/magistral/student/internal/service"
)

var (
	errNilStudentService = errors.New("student service is nil")
	errNilGroupService   = errors.New("group service is nil")
	errNilProgramService = errors.New("study program service is nil")
)

type api struct {
	studentService service.StudentService
	groupService   service.GroupService
	programService service.StudyProgramService
}

func New(studentService service.StudentService, groupService service.GroupService, programService service.StudyProgramService) (*api, error) {
	if studentService == nil {
		return nil, errNilStudentService
	}
	if groupService == nil {
		return nil, errNilGroupService
	}
	if programService == nil {
		return nil, errNilProgramService
	}
	return &api{
		studentService: studentService,
		groupService:   groupService,
		programService: programService,
	}, nil
}

func NewServer(studentService service.StudentService, groupService service.GroupService, programService service.StudyProgramService, opts ...studentsv1.ServerOption) (*studentsv1.Server, error) {
	handler, err := New(studentService, groupService, programService)
	if err != nil {
		return nil, err
	}

	return studentsv1.NewServer(handler, opts...)
}

func (a *api) schoolID(ctx context.Context) (string, error) {
	schoolID, ok := httpauth.SchoolIDFromContext(ctx)
	if !ok || schoolID == "" {
		return "", service.ErrInvalidArgument
	}
	return schoolID, nil
}
