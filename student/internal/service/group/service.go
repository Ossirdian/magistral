package group

import (
	"github.com/ossirdian/magistral/student/internal/repository"
	def "github.com/ossirdian/magistral/student/internal/service"
)

var _ def.GroupService = (*Service)(nil)

type Service struct {
	repo repository.GroupRepository
}

func NewService(repo repository.GroupRepository) *Service {
	return &Service{repo: repo}
}
