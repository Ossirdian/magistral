package api

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/resources/internal/model"
	repo "github.com/ossirdian/magistral/resources/internal/repository"
	servicemocks "github.com/ossirdian/magistral/resources/internal/service/mocks"
	resourcesv1 "github.com/ossirdian/magistral/shared/pkg/resources/v1"
	"github.com/stretchr/testify/suite"
)

type APISuite struct {
	suite.Suite

	ctx               context.Context
	schoolID          string
	carService        *servicemocks.CarService
	instructorService *servicemocks.InstructorService
	handler           *api
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func (s *APISuite) SetupTest() {
	s.ctx = context.Background()
	s.schoolID = "85233f17-b99f-4786-b0c3-aec780962f6d"
	s.carService = servicemocks.NewCarService(s.T())
	s.instructorService = servicemocks.NewInstructorService(s.T())

	var err error
	s.handler, err = New(s.carService, s.instructorService, s.schoolID)
	s.Require().NoError(err)
}

func (s *APISuite) TestNewValidatesDependencies() {
	handler, err := New(nil, s.instructorService, s.schoolID)
	s.Nil(handler)
	s.ErrorIs(err, errNilCarService)

	handler, err = New(s.carService, nil, s.schoolID)
	s.Nil(handler)
	s.ErrorIs(err, errNilInstructorService)

	handler, err = New(s.carService, s.instructorService, "")
	s.Nil(handler)
	s.ErrorIs(err, errEmptySchoolID)
}

func (s *APISuite) TestCreateCarReturnsConvertedResponse() {
	branchID := uuid.MustParse("5a0e8726-c9ec-40c0-83cc-8b904eef4ee2")
	instructorID := uuid.MustParse("8357f6c5-34f0-4c1a-84a0-7dc5cf039465")
	now := time.Unix(1700000000, 0).UTC()
	req := &resourcesv1.CarCreateRequest{
		BranchID:     branchID,
		InstructorID: resourcesv1.NewOptNilUUID(instructorID),
		PlateNumber:  "A123AA",
	}

	s.carService.EXPECT().
		Create(s.ctx, s.schoolID, model.CarCreate{
			BranchID:     branchID.String(),
			InstructorID: stringPtr(instructorID.String()),
			PlateNumber:  "A123AA",
			Brand:        nil,
			Model:        nil,
			Color:        nil,
			Status:       nil,
		}).
		Return(model.Car{
			ID:                 "9e2c8f55-66d4-426d-b47e-7cd7de2b4741",
			SchoolID:           s.schoolID,
			BranchID:           branchID.String(),
			InstructorID:       stringPtr(instructorID.String()),
			PlateNumber:        "A123AA",
			Status:             model.CarStatusActive,
			AvailabilityStatus: model.AvailabilityStatusAvailable,
			CreatedAt:          now,
			UpdatedAt:          now,
		}, nil).
		Once()

	res, err := s.handler.CreateCar(s.ctx, req)

	s.Require().NoError(err)
	response, ok := res.(*resourcesv1.CarResponse)
	s.Require().True(ok)
	s.Equal(uuid.MustParse("9e2c8f55-66d4-426d-b47e-7cd7de2b4741"), response.Data.ID)
	s.Equal(branchID, response.Data.BranchID)
	gotInstructorID, ok := response.Data.InstructorID.Get()
	s.True(ok)
	s.Equal(instructorID, gotInstructorID)
	s.Equal("A123AA", response.Data.PlateNumber)
}

func (s *APISuite) TestGetCarMapsNotFound() {
	carID := uuid.MustParse("9e2c8f55-66d4-426d-b47e-7cd7de2b4741")

	s.carService.EXPECT().
		Get(s.ctx, s.schoolID, carID.String()).
		Return(model.Car{}, repo.ErrNotFound).
		Once()

	res, err := s.handler.GetCar(s.ctx, resourcesv1.GetCarParams{CarId: carID})

	s.Require().NoError(err)
	notFound, ok := res.(*resourcesv1.GetCarNotFound)
	s.Require().True(ok)
	s.Equal("not_found", notFound.Error.Code)
	s.Equal("resource not found", notFound.Error.Message)
}

func (s *APISuite) TestListCarsReturnsDataAndMeta() {
	branchID := uuid.MustParse("5a0e8726-c9ec-40c0-83cc-8b904eef4ee2")
	carID := "9e2c8f55-66d4-426d-b47e-7cd7de2b4741"
	createdAt := time.Unix(1700000000, 0).UTC()

	var limit resourcesv1.OptInt
	limit.SetTo(25)
	var offset resourcesv1.OptInt
	offset.SetTo(10)
	var branchFilter resourcesv1.OptUUID
	branchFilter.SetTo(branchID)

	s.carService.EXPECT().
		List(s.ctx, s.schoolID, model.CarFilter{
			BranchID:           stringPtr(branchID.String()),
			InstructorID:       nil,
			Status:             nil,
			AvailabilityStatus: nil,
			Limit:              25,
			Offset:             10,
		}).
		Return([]model.Car{
			{
				ID:                 carID,
				SchoolID:           s.schoolID,
				BranchID:           branchID.String(),
				PlateNumber:        "A123AA",
				Status:             model.CarStatusActive,
				AvailabilityStatus: model.AvailabilityStatusAvailable,
				CreatedAt:          createdAt,
				UpdatedAt:          createdAt,
			},
		}, model.PageMeta{
			Limit:  25,
			Offset: 10,
			Count:  1,
		}, nil).
		Once()

	res, err := s.handler.ListCars(s.ctx, resourcesv1.ListCarsParams{
		BranchID: branchFilter,
		Limit:    limit,
		Offset:   offset,
	})

	s.Require().NoError(err)
	response, ok := res.(*resourcesv1.CarListResponse)
	s.Require().True(ok)
	s.Len(response.Data, 1)
	s.Equal(uuid.MustParse(carID), response.Data[0].ID)
	s.Equal(25, response.Meta.Limit)
	s.Equal(10, response.Meta.Offset)
	s.Equal(1, response.Meta.Count)
}

func (s *APISuite) TestCreateInstructorMapsConflict() {
	branchID := uuid.MustParse("5a0e8726-c9ec-40c0-83cc-8b904eef4ee2")
	req := &resourcesv1.InstructorCreateRequest{
		BranchID: branchID,
		FullName: "Ivan Petrov",
	}

	s.instructorService.EXPECT().
		Create(s.ctx, s.schoolID, model.InstructorCreate{
			BranchID: branchID.String(),
			FullName: "Ivan Petrov",
			Phone:    nil,
			Email:    nil,
			Status:   nil,
		}).
		Return(model.Instructor{}, repo.ErrConflict).
		Once()

	res, err := s.handler.CreateInstructor(s.ctx, req)

	s.Require().NoError(err)
	conflict, ok := res.(*resourcesv1.CreateInstructorConflict)
	s.Require().True(ok)
	s.Equal("conflict", conflict.Error.Code)
	s.Equal("resource conflict", conflict.Error.Message)
}

func stringPtr(value string) *string {
	return &value
}
