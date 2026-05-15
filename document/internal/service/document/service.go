package document

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/document/internal/model"
	"github.com/ossirdian/magistral/document/internal/repository"
	svc "github.com/ossirdian/magistral/document/internal/service"
	"github.com/ossirdian/magistral/document/internal/storage"
)

type StudentValidator interface {
	EnsureStudentBelongsToSchool(ctx context.Context, schoolID, studentID string) error
}

type Service struct {
	repo             repository.DocumentRepository
	storage          storage.Storage
	studentValidator StudentValidator
}

func NewService(repo repository.DocumentRepository, storage storage.Storage, studentValidator StudentValidator) *Service {
	return &Service{repo: repo, storage: storage, studentValidator: studentValidator}
}

func (s *Service) Upload(ctx context.Context, schoolID string, input model.DocumentUpload) (model.Document, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.Document{}, svc.ErrInvalidArgument
	}
	input = normalizeUpload(input)
	if input.StudentID == "" || input.Type == "" || input.Title == "" || input.MimeType == "" || len(input.Content) == 0 {
		return model.Document{}, svc.ErrInvalidArgument
	}
	if err := s.studentValidator.EnsureStudentBelongsToSchool(ctx, schoolID, input.StudentID); err != nil {
		return model.Document{}, err
	}

	documentID := uuid.NewString()
	objectKey := buildObjectKey(schoolID, input.StudentID, documentID, input.FileName, input.MimeType)
	if err := s.storage.Save(ctx, objectKey, input.Content); err != nil {
		return model.Document{}, err
	}

	item, err := s.repo.CreateDocument(ctx, schoolID, model.Document{
		ID:        documentID,
		StudentID: input.StudentID,
		Type:      input.Type,
		Title:     input.Title,
		Status:    model.DocumentStatusUploaded,
		ObjectKey: objectKey,
		MimeType:  input.MimeType,
		Size:      int64(len(input.Content)),
	})
	if err != nil {
		_ = s.storage.Delete(ctx, objectKey)
		return model.Document{}, err
	}
	return item, nil
}

func (s *Service) Generate(ctx context.Context, schoolID string, input model.DocumentGenerate) (model.Document, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.Document{}, svc.ErrInvalidArgument
	}
	input = normalizeGenerate(input)
	if input.StudentID == "" || input.TemplateID == "" || input.Title == "" {
		return model.Document{}, svc.ErrInvalidArgument
	}
	if err := s.studentValidator.EnsureStudentBelongsToSchool(ctx, schoolID, input.StudentID); err != nil {
		return model.Document{}, err
	}

	templateItem, err := s.repo.GetTemplate(ctx, schoolID, input.TemplateID)
	if err != nil {
		return model.Document{}, err
	}

	content, err := renderTemplate(templateItem.Content, input.Placeholders)
	if err != nil {
		return model.Document{}, svc.ErrInvalidArgument
	}

	documentID := uuid.NewString()
	fileName := input.Title + ".txt"
	if input.FileName != nil && strings.TrimSpace(*input.FileName) != "" {
		fileName = strings.TrimSpace(*input.FileName)
	}
	objectKey := buildObjectKey(schoolID, input.StudentID, documentID, fileName, "text/plain")
	if err := s.storage.Save(ctx, objectKey, content); err != nil {
		return model.Document{}, err
	}

	item, err := s.repo.CreateDocument(ctx, schoolID, model.Document{
		ID:        documentID,
		StudentID: input.StudentID,
		Type:      templateItem.Type,
		Title:     input.Title,
		Status:    model.DocumentStatusGenerated,
		ObjectKey: objectKey,
		MimeType:  "text/plain",
		Size:      int64(len(content)),
	})
	if err != nil {
		_ = s.storage.Delete(ctx, objectKey)
		return model.Document{}, err
	}
	return item, nil
}

func (s *Service) ListDocuments(ctx context.Context, schoolID string, filter model.DocumentFilter) ([]model.Document, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.StudentID != nil {
		value := strings.TrimSpace(*filter.StudentID)
		if value == "" {
			filter.StudentID = nil
		} else {
			filter.StudentID = &value
		}
	}
	if filter.Type != nil {
		value := strings.TrimSpace(*filter.Type)
		if value == "" {
			filter.Type = nil
		} else {
			filter.Type = &value
		}
	}
	if filter.Status != nil && !isValidStatus(*filter.Status) {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	return s.repo.ListDocuments(ctx, schoolID, filter)
}

func (s *Service) GetDocument(ctx context.Context, schoolID, documentID string) (model.Document, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(documentID) == "" {
		return model.Document{}, svc.ErrInvalidArgument
	}
	return s.repo.GetDocument(ctx, schoolID, documentID)
}

func (s *Service) Download(ctx context.Context, schoolID, documentID string) (model.DocumentDownload, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(documentID) == "" {
		return model.DocumentDownload{}, svc.ErrInvalidArgument
	}
	item, err := s.repo.GetDocument(ctx, schoolID, documentID)
	if err != nil {
		return model.DocumentDownload{}, err
	}
	content, err := s.storage.Load(ctx, item.ObjectKey)
	if err != nil {
		if err == storage.ErrNotFound {
			return model.DocumentDownload{}, repository.ErrNotFound
		}
		return model.DocumentDownload{}, err
	}
	return model.DocumentDownload{
		DocumentID: item.ID,
		FileName:   deriveFileName(item.Title, item.ObjectKey, item.MimeType),
		MimeType:   item.MimeType,
		Content:    content,
	}, nil
}

func (s *Service) DeleteDocument(ctx context.Context, schoolID, documentID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(documentID) == "" {
		return svc.ErrInvalidArgument
	}
	item, err := s.repo.GetDocument(ctx, schoolID, documentID)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(ctx, item.ObjectKey); err != nil {
		return err
	}
	return s.repo.DeleteDocument(ctx, schoolID, documentID)
}

func (s *Service) ListTemplates(ctx context.Context, schoolID string, filter model.DocumentTemplateFilter) ([]model.DocumentTemplate, model.PageMeta, error) {
	if strings.TrimSpace(schoolID) == "" {
		return nil, model.PageMeta{}, svc.ErrInvalidArgument
	}
	if filter.Type != nil {
		value := strings.TrimSpace(*filter.Type)
		if value == "" {
			filter.Type = nil
		} else {
			filter.Type = &value
		}
	}
	return s.repo.ListTemplates(ctx, schoolID, filter)
}

func (s *Service) CreateTemplate(ctx context.Context, schoolID string, input model.DocumentTemplateCreate) (model.DocumentTemplate, error) {
	if strings.TrimSpace(schoolID) == "" {
		return model.DocumentTemplate{}, svc.ErrInvalidArgument
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.TrimSpace(input.Type)
	input.Content = strings.TrimSpace(input.Content)
	if input.Name == "" || input.Type == "" || input.Content == "" {
		return model.DocumentTemplate{}, svc.ErrInvalidArgument
	}
	return s.repo.CreateTemplate(ctx, schoolID, input)
}

func (s *Service) UpdateTemplate(ctx context.Context, schoolID, templateID string, patch model.DocumentTemplatePatch) (model.DocumentTemplate, error) {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(templateID) == "" {
		return model.DocumentTemplate{}, svc.ErrInvalidArgument
	}
	patch = normalizeTemplatePatch(patch)
	if !patch.NameSet && !patch.TypeSet && !patch.ContentSet {
		return model.DocumentTemplate{}, svc.ErrInvalidArgument
	}
	return s.repo.UpdateTemplate(ctx, schoolID, templateID, patch)
}

func (s *Service) DeleteTemplate(ctx context.Context, schoolID, templateID string) error {
	if strings.TrimSpace(schoolID) == "" || strings.TrimSpace(templateID) == "" {
		return svc.ErrInvalidArgument
	}
	return s.repo.DeleteTemplate(ctx, schoolID, templateID)
}

func normalizeUpload(input model.DocumentUpload) model.DocumentUpload {
	input.StudentID = strings.TrimSpace(input.StudentID)
	input.Type = strings.TrimSpace(input.Type)
	input.Title = strings.TrimSpace(input.Title)
	input.MimeType = strings.TrimSpace(input.MimeType)
	input.FileName = strings.TrimSpace(input.FileName)
	return input
}

func normalizeGenerate(input model.DocumentGenerate) model.DocumentGenerate {
	input.StudentID = strings.TrimSpace(input.StudentID)
	input.TemplateID = strings.TrimSpace(input.TemplateID)
	input.Title = strings.TrimSpace(input.Title)
	if input.FileName != nil {
		value := strings.TrimSpace(*input.FileName)
		input.FileName = &value
	}
	if input.Placeholders == nil {
		input.Placeholders = map[string]string{}
	}
	return input
}

func normalizeTemplatePatch(patch model.DocumentTemplatePatch) model.DocumentTemplatePatch {
	if patch.NameSet && patch.Name != nil {
		value := strings.TrimSpace(*patch.Name)
		patch.Name = &value
	}
	if patch.TypeSet && patch.Type != nil {
		value := strings.TrimSpace(*patch.Type)
		patch.Type = &value
	}
	if patch.ContentSet && patch.Content != nil {
		value := strings.TrimSpace(*patch.Content)
		patch.Content = &value
	}
	return patch
}

func renderTemplate(content string, placeholders map[string]string) ([]byte, error) {
	tpl, err := template.New("document").Option("missingkey=error").Parse(content)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, placeholders); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildObjectKey(schoolID, studentID, documentID, fileName, mimeType string) string {
	ext := strings.TrimSpace(filepath.Ext(fileName))
	if ext == "" {
		extensions, _ := mime.ExtensionsByType(mimeType)
		if len(extensions) > 0 {
			ext = extensions[0]
		}
	}
	if ext == "" {
		ext = ".bin"
	}
	return fmt.Sprintf("schools/%s/students/%s/documents/%s%s", schoolID, studentID, documentID, ext)
}

func deriveFileName(title, objectKey, mimeType string) string {
	title = strings.TrimSpace(title)
	ext := filepath.Ext(objectKey)
	if ext == "" {
		extensions, _ := mime.ExtensionsByType(mimeType)
		if len(extensions) > 0 {
			ext = extensions[0]
		}
	}
	if ext == "" {
		ext = ".bin"
	}
	if title == "" {
		return "document" + ext
	}
	return title + ext
}

func isValidStatus(status model.DocumentStatus) bool {
	switch status {
	case model.DocumentStatusDraft, model.DocumentStatusGenerated, model.DocumentStatusUploaded, model.DocumentStatusArchived:
		return true
	default:
		return false
	}
}
