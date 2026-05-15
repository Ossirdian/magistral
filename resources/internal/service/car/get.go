package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) Get(ctx context.Context, schoolID, carID string) (model.Car, error) {
	return s.carRepo.GetByID(ctx, schoolID, carID)
}
