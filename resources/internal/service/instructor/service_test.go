package instructor

import (
	"context"
	"testing"
	"time"

	"github.com/ossirdian/magistral/resources/internal/model"
	repomocks "github.com/ossirdian/magistral/resources/internal/repository/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite

	ctx     context.Context
	repo    *repomocks.InstructorRepository
	service *Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctx = context.Background()
	s.repo = repomocks.NewInstructorRepository(s.T())
	s.service = NewService(s.repo)
}

func (s *ServiceSuite) TestCreateDelegatesToRepository() {
	input := model.InstructorCreate{
		BranchID: "branch-1",
		FullName: "Ivan Petrov",
	}
	expected := model.Instructor{
		ID:        "6bf0285f-5a41-4527-87d1-52b8cfd82957",
		SchoolID:  "3f1da7c5-c9db-4b34-899f-5fe4590d541c",
		BranchID:  input.BranchID,
		FullName:  input.FullName,
		Status:    model.InstructorStatusActive,
		CreatedAt: time.Unix(100, 0),
		UpdatedAt: time.Unix(200, 0),
	}

	s.repo.EXPECT().
		Create(s.ctx, "school-1", input).
		Return(expected, nil).
		Once()

	actual, err := s.service.Create(s.ctx, "school-1", input)

	s.Require().NoError(err)
	s.Equal(expected, actual)
}

func (s *ServiceSuite) TestListNormalizesFilterBeforeRepositoryCall() {
	repoItems := []model.Instructor{{ID: "instructor-1"}}
	repoMeta := model.PageMeta{Limit: 50, Offset: 0, Count: 1}

	s.repo.EXPECT().
		List(s.ctx, "school-1", mock.MatchedBy(func(filter model.InstructorFilter) bool {
			return filter.Limit == 50 && filter.Offset == 0
		})).
		Return(repoItems, repoMeta, nil).
		Once()

	actualItems, actualMeta, err := s.service.List(s.ctx, "school-1", model.InstructorFilter{
		Limit:  -5,
		Offset: -1,
	})

	s.Require().NoError(err)
	s.Equal(repoItems, actualItems)
	s.Equal(repoMeta, actualMeta)
}

func (s *ServiceSuite) TestGetDelegatesToRepository() {
	expected := model.Instructor{
		ID:       "instructor-1",
		SchoolID: "school-1",
		FullName: "Ivan Petrov",
	}

	s.repo.EXPECT().
		GetByID(s.ctx, "school-1", "instructor-1").
		Return(expected, nil).
		Once()

	actual, err := s.service.Get(s.ctx, "school-1", "instructor-1")

	s.Require().NoError(err)
	s.Equal(expected, actual)
}
