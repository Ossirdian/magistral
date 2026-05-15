package program

import (
	"context"
	"strings"

	"github.com/ossirdian/magistral/student/internal/model"
	"github.com/ossirdian/magistral/student/internal/repository"
	def "github.com/ossirdian/magistral/student/internal/service"
)

type Service struct {
	repo repository.StudyProgramRepository
}

func NewService(repo repository.StudyProgramRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, schoolID string, input model.StudyProgramCreate) (model.StudyProgram, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.StudyProgram{}, def.ErrInvalidArgument
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.TotalHours < 0 || !isValidTransmission(input.Transmission) {
		return model.StudyProgram{}, def.ErrInvalidArgument
	}
	return s.repo.Create(ctx, schoolID, input)
}

func (s *Service) Get(ctx context.Context, schoolID, programID string) (model.StudyProgram, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(programID) == "" {
		return model.StudyProgram{}, def.ErrInvalidArgument
	}
	return s.repo.GetByID(ctx, schoolID, programID)
}

func (s *Service) List(ctx context.Context, schoolID string, filter model.StudyProgramFilter) ([]model.StudyProgram, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, def.ErrInvalidArgument
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.List(ctx, schoolID, filter)
}

func (s *Service) Update(ctx context.Context, schoolID, programID string, patch model.StudyProgramPatch) (model.StudyProgram, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(programID) == "" {
		return model.StudyProgram{}, def.ErrInvalidArgument
	}
	if patch.Name != nil {
		value := strings.TrimSpace(*patch.Name)
		if value == "" {
			return model.StudyProgram{}, def.ErrInvalidArgument
		}
		patch.Name = &value
	}
	if patch.Transmission != nil && !isValidTransmission(*patch.Transmission) {
		return model.StudyProgram{}, def.ErrInvalidArgument
	}
	if patch.TotalHours != nil && *patch.TotalHours < 0 {
		return model.StudyProgram{}, def.ErrInvalidArgument
	}
	return s.repo.Update(ctx, schoolID, programID, patch)
}

func (s *Service) Delete(ctx context.Context, schoolID, programID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(programID) == "" {
		return def.ErrInvalidArgument
	}
	return s.repo.Delete(ctx, schoolID, programID)
}

func isValidTransmission(value model.TransmissionType) bool {
	switch value {
	case model.TransmissionTypeManual, model.TransmissionTypeAutomatic:
		return true
	default:
		return false
	}
}

var _ def.StudyProgramService = (*Service)(nil)
