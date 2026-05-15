package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) GetAvailability(ctx context.Context, schoolID, carID string) ([]model.AvailabilityWindow, error) {
	return s.carRepo.GetAvailability(ctx, schoolID, carID)
}

func (s *Service) ReplaceAvailability(ctx context.Context, schoolID, carID string, windows []model.AvailabilityWindowInput) ([]model.AvailabilityWindow, error) {
	return s.carRepo.ReplaceAvailability(ctx, schoolID, carID, windows)
}
