package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) Create(ctx context.Context, schoolID string, input model.CarCreate) (model.Car, error) {
	return s.carRepo.Create(ctx, schoolID, input)
}
