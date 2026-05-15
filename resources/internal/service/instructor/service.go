package instructor

import (
	"github.com/ossirdian/magistral/resources/internal/repository"
	def "github.com/ossirdian/magistral/resources/internal/service"
)

var _ def.InstructorService = (*Service)(nil)

type Service struct {
	instructorRepo repository.InstructorRepository
}

func NewService(instructorRepo repository.InstructorRepository) *Service {
	return &Service{instructorRepo: instructorRepo}
}
