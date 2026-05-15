package car

import "context"

func (s *Service) Delete(ctx context.Context, schoolID, carID string) error {
	return s.carRepo.Delete(ctx, schoolID, carID)
}
