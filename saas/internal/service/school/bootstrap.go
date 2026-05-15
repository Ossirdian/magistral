package school

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ossirdian/magistral/saas/internal/model"
	def "github.com/ossirdian/magistral/saas/internal/service"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

var ErrIAMClientUnavailable = errors.New("iam client is unavailable")

func (s *Service) BootstrapWithAdmin(ctx context.Context, input model.SchoolCreate, admin model.SchoolAdminBootstrap) (model.School, string, error) {
	if s.iamClient == nil {
		return model.School{}, "", ErrIAMClientUnavailable
	}
	if err := validateAdminBootstrap(admin); err != nil {
		return model.School{}, "", err
	}

	school, err := s.schoolRepo.Create(ctx, input)
	if err != nil {
		return model.School{}, "", err
	}

	adminUserID, err := s.iamClient.CreateSchoolAdmin(ctx, school.ID, admin.Email, admin.Password)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to create school admin, rolling back school", "school_id", school.ID, "error", err)
		rollbackErr := s.schoolRepo.Delete(ctx, school.ID)
		if rollbackErr != nil {
			s.logger.WithContext(ctx).Error("failed to rollback school after admin creation failure", "school_id", school.ID, "rollback_error", rollbackErr)
			return model.School{}, "", fmt.Errorf("failed to create admin: %w; rollback failed: %v", err, rollbackErr)
		}
		if grpcstatus.Code(err) == codes.InvalidArgument {
			return model.School{}, "", def.ErrInvalidArgument
		}
		return model.School{}, "", err
	}

	return school, adminUserID, nil
}

func validateAdminBootstrap(admin model.SchoolAdminBootstrap) error {
	email := strings.TrimSpace(admin.Email)
	password := admin.Password

	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return def.ErrInvalidArgument
	}
	if len(password) < 8 {
		return def.ErrInvalidArgument
	}

	hasLetter, hasDigit := false, false
	for _, r := range password {
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return def.ErrInvalidArgument
	}

	return nil
}
