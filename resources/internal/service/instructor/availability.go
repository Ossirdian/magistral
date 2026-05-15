package instructor

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) GetAvailability(ctx context.Context, schoolID, instructorID string) ([]model.AvailabilityWindow, error) {
	return s.instructorRepo.GetAvailability(ctx, schoolID, instructorID)
}

func (s *Service) ReplaceAvailability(ctx context.Context, schoolID, instructorID string, windows []model.AvailabilityWindowInput) ([]model.AvailabilityWindow, error) {
	return s.instructorRepo.ReplaceAvailability(ctx, schoolID, instructorID, windows)
}
