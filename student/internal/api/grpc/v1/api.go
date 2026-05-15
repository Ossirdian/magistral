package v1

import (
	"context"
	"errors"

	studentv1 "github.com/ossirdian/magistral/shared/pkg/student/v1"
	repo "github.com/ossirdian/magistral/student/internal/repository"
	svc "github.com/ossirdian/magistral/student/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type api struct {
	studentv1.UnimplementedStudentServiceServer
	service svc.StudentService
}

func New(service svc.StudentService) (*api, error) {
	if service == nil {
		return nil, errors.New("student service is nil")
	}
	return &api{service: service}, nil
}

func (a *api) GetStudent(ctx context.Context, req *studentv1.GetStudentRequest) (*studentv1.Student, error) {
	item, err := a.service.Get(ctx, req.GetSchoolId(), req.GetStudentId())
	if err != nil {
		return nil, mapError(err)
	}

	result := &studentv1.Student{
		Id:        item.ID,
		SchoolId:  item.SchoolID,
		BranchId:  item.BranchID,
		FullName:  item.FullName,
		Phone:     item.Phone,
		Status:    statusToProto(string(item.Status)),
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}
	if item.GroupID != nil {
		result.GroupId = *item.GroupID
	}
	return result, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, svc.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, repo.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, repo.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func statusToProto(value string) studentv1.StudentStatus {
	switch value {
	case "new":
		return studentv1.StudentStatus_STUDENT_STATUS_NEW
	case "active":
		return studentv1.StudentStatus_STUDENT_STATUS_ACTIVE
	case "paused":
		return studentv1.StudentStatus_STUDENT_STATUS_PAUSED
	case "graduated":
		return studentv1.StudentStatus_STUDENT_STATUS_GRADUATED
	case "archived":
		return studentv1.StudentStatus_STUDENT_STATUS_ARCHIVED
	default:
		return studentv1.StudentStatus_STUDENT_STATUS_UNSPECIFIED
	}
}

var _ studentv1.StudentServiceServer = (*api)(nil)
