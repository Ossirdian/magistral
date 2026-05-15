package repository

import (
	"context"

	"github.com/ossirdian/magistral/document/internal/model"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, schoolID string, item model.Document) (model.Document, error)
	GetDocument(ctx context.Context, schoolID, documentID string) (model.Document, error)
	ListDocuments(ctx context.Context, schoolID string, filter model.DocumentFilter) ([]model.Document, model.PageMeta, error)
	DeleteDocument(ctx context.Context, schoolID, documentID string) error
	GetTemplate(ctx context.Context, schoolID, templateID string) (model.DocumentTemplate, error)
	ListTemplates(ctx context.Context, schoolID string, filter model.DocumentTemplateFilter) ([]model.DocumentTemplate, model.PageMeta, error)
	CreateTemplate(ctx context.Context, schoolID string, input model.DocumentTemplateCreate) (model.DocumentTemplate, error)
	UpdateTemplate(ctx context.Context, schoolID, templateID string, patch model.DocumentTemplatePatch) (model.DocumentTemplate, error)
	DeleteTemplate(ctx context.Context, schoolID, templateID string) error
}
