package student

import (
	"context"

	"github.com/ossirdian/magistral/student/internal/repository"
	def "github.com/ossirdian/magistral/student/internal/service"
)

var _ def.StudentService = (*Service)(nil)

type Service struct {
	repo           repository.StudentRepository
	programs       repository.StudyProgramRepository
	lessonProgress lessonProgressReader
}

type lessonProgressReader interface {
	CompletedHoursByStudentIDs(ctx context.Context, schoolID string, studentIDs []string) (map[string]float64, error)
}

func NewService(repo repository.StudentRepository, programs repository.StudyProgramRepository, lessonProgress lessonProgressReader) *Service {
	return &Service{repo: repo, programs: programs, lessonProgress: lessonProgress}
}
