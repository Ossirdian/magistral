package student

import (
	"context"

	repo "github.com/ossirdian/magistral/document/internal/repository"
	studentv1 "github.com/ossirdian/magistral/shared/pkg/student/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client studentv1.StudentServiceClient
}

func New(client studentv1.StudentServiceClient) *Client {
	if client == nil {
		return nil
	}
	return &Client{client: client}
}

func (c *Client) EnsureStudentBelongsToSchool(ctx context.Context, schoolID, studentID string) error {
	_, err := c.client.GetStudent(ctx, &studentv1.GetStudentRequest{
		SchoolId:  schoolID,
		StudentId: studentID,
	})
	if err == nil {
		return nil
	}

	switch status.Code(err) {
	case codes.NotFound:
		return repo.ErrNotFound
	case codes.InvalidArgument:
		return repo.ErrConflict
	default:
		return err
	}
}
