package branch

import "context"

func (s *Service) Delete(ctx context.Context, branchID string) error {
	return s.branchRepo.Delete(ctx, branchID)
}
