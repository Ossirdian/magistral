package lesson

import (
	"github.com/ossirdian/magistral/lesson/internal/repository"
	def "github.com/ossirdian/magistral/lesson/internal/service"
)

var _ def.LessonService = (*Service)(nil)

type Service struct {
	repo repository.LessonRepository
}

func NewService(repo repository.LessonRepository) *Service {
	return &Service{repo: repo}
}
