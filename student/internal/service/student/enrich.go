package student

import (
	"context"

	"github.com/ossirdian/magistral/student/internal/model"
)

func (s *Service) enrichStudent(ctx context.Context, schoolID string, item model.Student) (model.Student, error) {
	items, err := s.enrichStudents(ctx, schoolID, []model.Student{item})
	if err != nil {
		return model.Student{}, err
	}
	return items[0], nil
}

func (s *Service) enrichStudents(ctx context.Context, schoolID string, items []model.Student) ([]model.Student, error) {
	if len(items) == 0 {
		return items, nil
	}

	programHours := make(map[string]float64)
	if s.programs != nil {
		programs, _, err := s.programs.List(ctx, schoolID, model.StudyProgramFilter{Limit: 1000})
		if err != nil {
			return nil, err
		}
		for _, program := range programs {
			programHours[program.ID] = program.TotalHours
		}
	}

	completedByStudent := make(map[string]float64)
	if s.lessonProgress != nil {
		studentIDs := make([]string, 0, len(items))
		for _, item := range items {
			studentIDs = append(studentIDs, item.ID)
		}
		progress, err := s.lessonProgress.CompletedHoursByStudentIDs(ctx, schoolID, studentIDs)
		if err != nil {
			return nil, err
		}
		completedByStudent = progress
	}

	for i := range items {
		items[i].LessonHoursCompleted = completedByStudent[items[i].ID]
		items[i].TotalHoursCompleted = items[i].ManualHoursCompleted + items[i].LessonHoursCompleted
		if items[i].ProgramID == nil {
			continue
		}
		totalHours, ok := programHours[*items[i].ProgramID]
		if !ok {
			continue
		}
		items[i].RemainingHours = totalHours - items[i].TotalHoursCompleted
		if items[i].RemainingHours < 0 {
			items[i].RemainingHours = 0
		}
	}

	return items, nil
}
