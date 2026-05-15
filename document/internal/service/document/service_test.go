package document

import (
	"context"
	"errors"
	"testing"

	"github.com/ossirdian/magistral/document/internal/model"
	"github.com/ossirdian/magistral/document/internal/repository"
	svc "github.com/ossirdian/magistral/document/internal/service"
	"github.com/ossirdian/magistral/document/internal/storage"
)

type studentValidatorStub struct {
	calls       int
	lastSchool  string
	lastStudent string
	err         error
}

func (s *studentValidatorStub) EnsureStudentBelongsToSchool(_ context.Context, schoolID, studentID string) error {
	s.calls++
	s.lastSchool = schoolID
	s.lastStudent = studentID
	return s.err
}

type storageStub struct {
	savedKey     string
	savedContent []byte
	loadedKey    string
	deleteKey    string
	loadContent  []byte
	saveErr      error
	loadErr      error
	deleteErr    error
}

func (s *storageStub) Save(_ context.Context, objectKey string, content []byte) error {
	s.savedKey = objectKey
	s.savedContent = append([]byte(nil), content...)
	return s.saveErr
}

func (s *storageStub) Load(_ context.Context, objectKey string) ([]byte, error) {
	s.loadedKey = objectKey
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return append([]byte(nil), s.loadContent...), nil
}

func (s *storageStub) Delete(_ context.Context, objectKey string) error {
	s.deleteKey = objectKey
	return s.deleteErr
}

type repositoryStub struct {
	createDocumentFn func(context.Context, string, model.Document) (model.Document, error)
	getDocumentFn    func(context.Context, string, string) (model.Document, error)
	listDocumentsFn  func(context.Context, string, model.DocumentFilter) ([]model.Document, model.PageMeta, error)
	deleteDocumentFn func(context.Context, string, string) error
	getTemplateFn    func(context.Context, string, string) (model.DocumentTemplate, error)
	listTemplatesFn  func(context.Context, string, model.DocumentTemplateFilter) ([]model.DocumentTemplate, model.PageMeta, error)
	createTemplateFn func(context.Context, string, model.DocumentTemplateCreate) (model.DocumentTemplate, error)
	updateTemplateFn func(context.Context, string, string, model.DocumentTemplatePatch) (model.DocumentTemplate, error)
	deleteTemplateFn func(context.Context, string, string) error
}

func (s *repositoryStub) CreateDocument(ctx context.Context, schoolID string, item model.Document) (model.Document, error) {
	if s.createDocumentFn == nil {
		panic("unexpected CreateDocument call")
	}
	return s.createDocumentFn(ctx, schoolID, item)
}

func (s *repositoryStub) GetDocument(ctx context.Context, schoolID, documentID string) (model.Document, error) {
	if s.getDocumentFn == nil {
		panic("unexpected GetDocument call")
	}
	return s.getDocumentFn(ctx, schoolID, documentID)
}

func (s *repositoryStub) ListDocuments(ctx context.Context, schoolID string, filter model.DocumentFilter) ([]model.Document, model.PageMeta, error) {
	if s.listDocumentsFn == nil {
		panic("unexpected ListDocuments call")
	}
	return s.listDocumentsFn(ctx, schoolID, filter)
}

func (s *repositoryStub) DeleteDocument(ctx context.Context, schoolID, documentID string) error {
	if s.deleteDocumentFn == nil {
		panic("unexpected DeleteDocument call")
	}
	return s.deleteDocumentFn(ctx, schoolID, documentID)
}

func (s *repositoryStub) GetTemplate(ctx context.Context, schoolID, templateID string) (model.DocumentTemplate, error) {
	if s.getTemplateFn == nil {
		panic("unexpected GetTemplate call")
	}
	return s.getTemplateFn(ctx, schoolID, templateID)
}

func (s *repositoryStub) ListTemplates(ctx context.Context, schoolID string, filter model.DocumentTemplateFilter) ([]model.DocumentTemplate, model.PageMeta, error) {
	if s.listTemplatesFn == nil {
		panic("unexpected ListTemplates call")
	}
	return s.listTemplatesFn(ctx, schoolID, filter)
}

func (s *repositoryStub) CreateTemplate(ctx context.Context, schoolID string, input model.DocumentTemplateCreate) (model.DocumentTemplate, error) {
	if s.createTemplateFn == nil {
		panic("unexpected CreateTemplate call")
	}
	return s.createTemplateFn(ctx, schoolID, input)
}

func (s *repositoryStub) UpdateTemplate(ctx context.Context, schoolID, templateID string, patch model.DocumentTemplatePatch) (model.DocumentTemplate, error) {
	if s.updateTemplateFn == nil {
		panic("unexpected UpdateTemplate call")
	}
	return s.updateTemplateFn(ctx, schoolID, templateID, patch)
}

func (s *repositoryStub) DeleteTemplate(ctx context.Context, schoolID, templateID string) error {
	if s.deleteTemplateFn == nil {
		panic("unexpected DeleteTemplate call")
	}
	return s.deleteTemplateFn(ctx, schoolID, templateID)
}

func TestUploadStoresContentAndCreatesMetadata(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		createDocumentFn: func(_ context.Context, schoolID string, item model.Document) (model.Document, error) {
			if schoolID != "school-1" {
				t.Fatalf("unexpected schoolID: %s", schoolID)
			}
			if item.StudentID != "student-1" || item.Status != model.DocumentStatusUploaded {
				t.Fatalf("unexpected document payload: %+v", item)
			}
			if item.Size != 4 {
				t.Fatalf("unexpected size: %d", item.Size)
			}
			return item, nil
		},
	}
	storage := &storageStub{}
	validator := &studentValidatorStub{}
	service := NewService(repo, storage, validator)

	item, err := service.Upload(context.Background(), "school-1", model.DocumentUpload{
		StudentID: "student-1",
		Type:      "contract",
		Title:     "Agreement",
		MimeType:  "application/pdf",
		FileName:  "agreement.pdf",
		Content:   []byte("file"),
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("expected validator call count 1, got %d", validator.calls)
	}
	if storage.savedKey == "" || len(storage.savedContent) != 4 {
		t.Fatalf("storage save was not called as expected: %+v", storage)
	}
	if item.ObjectKey == "" {
		t.Fatal("expected object key to be set")
	}
}

func TestGenerateRendersTemplateAndStoresDocument(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		getTemplateFn: func(_ context.Context, schoolID, templateID string) (model.DocumentTemplate, error) {
			if schoolID != "school-1" || templateID != "template-1" {
				t.Fatalf("unexpected template lookup: %s %s", schoolID, templateID)
			}
			return model.DocumentTemplate{
				ID:       "template-1",
				SchoolID: schoolID,
				Type:     "application",
				Content:  "Hello, {{.name}}!",
			}, nil
		},
		createDocumentFn: func(_ context.Context, _ string, item model.Document) (model.Document, error) {
			if item.Status != model.DocumentStatusGenerated {
				t.Fatalf("unexpected status: %s", item.Status)
			}
			return item, nil
		},
	}
	storage := &storageStub{}
	service := NewService(repo, storage, &studentValidatorStub{})

	_, err := service.Generate(context.Background(), "school-1", model.DocumentGenerate{
		StudentID:  "student-1",
		TemplateID: "template-1",
		Title:      "Generated",
		Placeholders: map[string]string{
			"name": "Vlad",
		},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if string(storage.savedContent) != "Hello, Vlad!" {
		t.Fatalf("unexpected generated content: %q", string(storage.savedContent))
	}
}

func TestDownloadMapsStorageNotFoundToRepositoryNotFound(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		getDocumentFn: func(_ context.Context, _, _ string) (model.Document, error) {
			return model.Document{
				ID:        "document-1",
				Title:     "Agreement",
				ObjectKey: "schools/s/documents/d.pdf",
				MimeType:  "application/pdf",
			}, nil
		},
	}
	storage := &storageStub{loadErr: storage.ErrNotFound}
	service := NewService(repo, storage, &studentValidatorStub{})

	_, err := service.Download(context.Background(), "school-1", "document-1")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteDocumentDeletesStorageAndMetadata(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{
		getDocumentFn: func(_ context.Context, _, _ string) (model.Document, error) {
			return model.Document{ID: "document-1", ObjectKey: "path/file.pdf"}, nil
		},
		deleteDocumentFn: func(_ context.Context, schoolID, documentID string) error {
			if schoolID != "school-1" || documentID != "document-1" {
				t.Fatalf("unexpected delete args: %s %s", schoolID, documentID)
			}
			return nil
		},
	}
	storage := &storageStub{}
	service := NewService(repo, storage, &studentValidatorStub{})

	if err := service.DeleteDocument(context.Background(), "school-1", "document-1"); err != nil {
		t.Fatalf("DeleteDocument returned error: %v", err)
	}
	if storage.deleteKey != "path/file.pdf" {
		t.Fatalf("unexpected deleted key: %s", storage.deleteKey)
	}
}

func TestUpdateTemplateRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	service := NewService(&repositoryStub{}, &storageStub{}, &studentValidatorStub{})

	_, err := service.UpdateTemplate(context.Background(), "school-1", "template-1", model.DocumentTemplatePatch{})
	if !errors.Is(err, svc.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
