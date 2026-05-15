package school

import "context"

func (s *Service) Delete(ctx context.Context, schoolID string) error {
	return s.schoolRepo.Delete(ctx, schoolID)
}
