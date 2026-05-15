package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) List(ctx context.Context, schoolID string, filter model.CarFilter) ([]model.Car, model.PageMeta, error) {
	filter = normalizeCarFilter(filter)
	return s.carRepo.List(ctx, schoolID, filter)
}
