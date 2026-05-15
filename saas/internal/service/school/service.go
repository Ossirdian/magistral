package school

import (
	"context"

	"github.com/ossirdian/magistral/platform/pkg/logger"
	"github.com/ossirdian/magistral/saas/internal/repository"
	def "github.com/ossirdian/magistral/saas/internal/service"
)

var _ def.SchoolService = (*Service)(nil)

type Service struct {
	schoolRepo repository.SchoolRepository
	iamClient  IAMClient
	logger     logger.Logger
}

type IAMClient interface {
	CreateSchoolAdmin(ctx context.Context, schoolID, email, password string) (string, error)
}

func NewService(schoolRepo repository.SchoolRepository, iamClient IAMClient, log logger.Logger) *Service {
	if log == nil {
		log = logger.Global()
	}

	return &Service{
		schoolRepo: schoolRepo,
		iamClient:  iamClient,
		logger:     log,
	}
}
