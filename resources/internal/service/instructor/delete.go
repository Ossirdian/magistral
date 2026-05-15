package instructor

import "context"

func (s *Service) Delete(ctx context.Context, schoolID, instructorID string) error {
	return s.instructorRepo.Delete(ctx, schoolID, instructorID)
}
