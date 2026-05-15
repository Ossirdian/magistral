package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) Update(ctx context.Context, schoolID, carID string, patch model.CarPatch) (model.Car, error) {
	return s.carRepo.Update(ctx, schoolID, carID, patch)
}
