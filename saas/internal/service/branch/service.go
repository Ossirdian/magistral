package branch

import (
	"github.com/ossirdian/magistral/saas/internal/repository"
	def "github.com/ossirdian/magistral/saas/internal/service"
)

var _ def.BranchService = (*Service)(nil)

type Service struct {
	branchRepo repository.BranchRepository
}

func NewService(branchRepo repository.BranchRepository) *Service {
	return &Service{branchRepo: branchRepo}
}
