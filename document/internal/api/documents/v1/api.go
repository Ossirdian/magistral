package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/document/internal/httpauth"
	"github.com/ossirdian/magistral/document/internal/service"
	documentsv1 "github.com/ossirdian/magistral/shared/pkg/documents/v1"
)

var errNilDocumentService = errors.New("document service is nil")

type api struct {
	documentService service.DocumentService
}

func New(documentService service.DocumentService) (*api, error) {
	if documentService == nil {
		return nil, errNilDocumentService
	}
	return &api{documentService: documentService}, nil
}

func NewServer(documentService service.DocumentService, opts ...documentsv1.ServerOption) (*documentsv1.Server, error) {
	handler, err := New(documentService)
	if err != nil {
		return nil, err
	}
	return documentsv1.NewServer(handler, opts...)
}

func (a *api) schoolID(ctx context.Context) (string, error) {
	schoolID, ok := httpauth.SchoolIDFromContext(ctx)
	if !ok || schoolID == "" {
		return "", service.ErrInvalidArgument
	}
	return schoolID, nil
}
