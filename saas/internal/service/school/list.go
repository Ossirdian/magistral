package school

import (
	"context"

	"github.com/ossirdian/magistral/saas/internal/model"
)

func (s *Service) List(ctx context.Context, filter model.SchoolFilter) ([]model.School, model.PageMeta, error) {
	filter = normalizeSchoolFilter(filter)
	return s.schoolRepo.List(ctx, filter)
}
