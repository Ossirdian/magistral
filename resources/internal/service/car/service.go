package car

import (
	"github.com/ossirdian/magistral/resources/internal/repository"
	def "github.com/ossirdian/magistral/resources/internal/service"
)

var _ def.CarService = (*Service)(nil)

type Service struct {
	carRepo repository.CarRepository
}

func NewService(carRepo repository.CarRepository) *Service {
	return &Service{carRepo: carRepo}
}
