package car

import (
	"context"

	"github.com/ossirdian/magistral/resources/internal/model"
)

func (s *Service) AssignInstructor(ctx context.Context, schoolID, carID string, input model.CarAssignment) (model.CarAssignmentResult, error) {
	return s.carRepo.AssignInstructor(ctx, schoolID, carID, input)
}
