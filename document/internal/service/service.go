package service

import (
	"context"

	"github.com/ossirdian/magistral/document/internal/model"
)

type DocumentService interface {
	Upload(ctx context.Context, schoolID string, input model.DocumentUpload) (model.Document, error)
	Generate(ctx context.Context, schoolID string, input model.DocumentGenerate) (model.Document, error)
	ListDocuments(ctx context.Context, schoolID string, filter model.DocumentFilter) ([]model.Document, model.PageMeta, error)
	GetDocument(ctx context.Context, schoolID, documentID string) (model.Document, error)
	Download(ctx context.Context, schoolID, documentID string) (model.DocumentDownload, error)
	DeleteDocument(ctx context.Context, schoolID, documentID string) error
	ListTemplates(ctx context.Context, schoolID string, filter model.DocumentTemplateFilter) ([]model.DocumentTemplate, model.PageMeta, error)
	CreateTemplate(ctx context.Context, schoolID string, input model.DocumentTemplateCreate) (model.DocumentTemplate, error)
	UpdateTemplate(ctx context.Context, schoolID, templateID string, patch model.DocumentTemplatePatch) (model.DocumentTemplate, error)
	DeleteTemplate(ctx context.Context, schoolID, templateID string) error
}
