package car

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
	repo    *repomocks.CarRepository
	service *Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctx = context.Background()
	s.repo = repomocks.NewCarRepository(s.T())
	s.service = NewService(s.repo)
}

func (s *ServiceSuite) TestCreateDelegatesToRepository() {
	input := model.CarCreate{
		BranchID:    "branch-1",
		PlateNumber: "A123AA",
	}
	expected := model.Car{
		ID:          "2a4e0523-b5d8-4d0c-8a3c-827eebac0eb1",
		SchoolID:    "85233f17-b99f-4786-b0c3-aec780962f6d",
		BranchID:    input.BranchID,
		PlateNumber: input.PlateNumber,
		Status:      model.CarStatusActive,
		CreatedAt:   time.Unix(100, 0),
		UpdatedAt:   time.Unix(200, 0),
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
	repoItems := []model.Car{{ID: "car-1"}}
	repoMeta := model.PageMeta{Limit: 50, Offset: 0, Count: 1}

	s.repo.EXPECT().
		List(s.ctx, "school-1", mock.MatchedBy(func(filter model.CarFilter) bool {
			return filter.Limit == 50 && filter.Offset == 0
		})).
		Return(repoItems, repoMeta, nil).
		Once()

	actualItems, actualMeta, err := s.service.List(s.ctx, "school-1", model.CarFilter{
		Limit:  0,
		Offset: -10,
	})

	s.Require().NoError(err)
	s.Equal(repoItems, actualItems)
	s.Equal(repoMeta, actualMeta)
}

func (s *ServiceSuite) TestAssignInstructorDelegatesToRepository() {
	input := model.CarAssignment{InstructorID: "instructor-1"}
	expected := model.CarAssignmentResult{
		CarID:        "car-1",
		InstructorID: "instructor-1",
		AssignedAt:   time.Unix(300, 0),
	}

	s.repo.EXPECT().
		AssignInstructor(s.ctx, "school-1", "car-1", input).
		Return(expected, nil).
		Once()

	actual, err := s.service.AssignInstructor(s.ctx, "school-1", "car-1", input)

	s.Require().NoError(err)
	s.Equal(expected, actual)
}
