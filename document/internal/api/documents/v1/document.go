package api

import (
	"context"
	"errors"

	"github.com/ossirdian/magistral/document/internal/converter"
	repo "github.com/ossirdian/magistral/document/internal/repository"
	svc "github.com/ossirdian/magistral/document/internal/service"
	documentsv1 "github.com/ossirdian/magistral/shared/pkg/documents/v1"
)

func (a *api) UploadDocument(ctx context.Context, req *documentsv1.DocumentUploadRequest, _ documentsv1.UploadDocumentParams) (documentsv1.UploadDocumentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.UploadDocumentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	input, err := converter.ConvertUploadRequest(req)
	if err != nil {
		res := documentsv1.UploadDocumentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
		return &res, nil
	}
	item, err := a.documentService.Upload(ctx, schoolID, input)
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := documentsv1.UploadDocumentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := documentsv1.UploadDocumentNotFound(newErrorBody(ctx, "not_found", "student not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := documentsv1.UploadDocumentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertDocument(item)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentResponse{Data: data}, nil
}

func (a *api) GenerateDocument(ctx context.Context, req *documentsv1.DocumentGenerateRequest, _ documentsv1.GenerateDocumentParams) (documentsv1.GenerateDocumentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.GenerateDocumentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.documentService.Generate(ctx, schoolID, converter.ConvertGenerateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := documentsv1.GenerateDocumentBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := documentsv1.GenerateDocumentNotFound(newErrorBody(ctx, "not_found", "resource not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := documentsv1.GenerateDocumentConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertDocument(item)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentResponse{Data: data}, nil
}

func (a *api) ListDocuments(ctx context.Context, params documentsv1.ListDocumentsParams) (documentsv1.ListDocumentsRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.ListDocumentsUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.documentService.ListDocuments(ctx, schoolID, converter.ConvertDocumentFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := documentsv1.ListDocumentsBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertDocuments(items)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentListResponse{Data: data, Meta: converter.ConvertMeta(meta)}, nil
}

func (a *api) GetDocument(ctx context.Context, params documentsv1.GetDocumentParams) (documentsv1.GetDocumentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.GetDocumentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.documentService.GetDocument(ctx, schoolID, params.DocumentId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := documentsv1.GetDocumentNotFound(newErrorBody(ctx, "not_found", "document not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertDocument(item)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentResponse{Data: data}, nil
}

func (a *api) DownloadDocument(ctx context.Context, params documentsv1.DownloadDocumentParams) (documentsv1.DownloadDocumentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.DownloadDocumentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.documentService.Download(ctx, schoolID, params.DocumentId.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := documentsv1.DownloadDocumentNotFound(newErrorBody(ctx, "not_found", "document not found"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertDownload(item)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentDownloadResponse{Data: data}, nil
}

func (a *api) DeleteDocument(ctx context.Context, params documentsv1.DeleteDocumentParams) (documentsv1.DeleteDocumentRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.DeleteDocumentUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.documentService.DeleteDocument(ctx, schoolID, params.DocumentId.String()); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := documentsv1.DeleteDocumentNotFound(newErrorBody(ctx, "not_found", "document not found"))
			return &res, nil
		}
		return nil, err
	}
	return &documentsv1.DeleteDocumentNoContent{}, nil
}

func (a *api) ListDocumentTemplates(ctx context.Context, params documentsv1.ListDocumentTemplatesParams) (documentsv1.ListDocumentTemplatesRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.ListDocumentTemplatesUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	items, meta, err := a.documentService.ListTemplates(ctx, schoolID, converter.ConvertTemplateFilter(params))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidArgument) {
			res := documentsv1.ListDocumentTemplatesBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		}
		return nil, err
	}

	data, err := converter.ConvertDocumentTemplates(items)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentTemplateListResponse{Data: data, Meta: converter.ConvertMeta(meta)}, nil
}

func (a *api) CreateDocumentTemplate(ctx context.Context, req *documentsv1.DocumentTemplateCreateRequest, _ documentsv1.CreateDocumentTemplateParams) (documentsv1.CreateDocumentTemplateRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.CreateDocumentTemplateUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.documentService.CreateTemplate(ctx, schoolID, converter.ConvertTemplateCreateRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := documentsv1.CreateDocumentTemplateBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := documentsv1.CreateDocumentTemplateConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertDocumentTemplate(item)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentTemplateResponse{Data: data}, nil
}

func (a *api) PatchDocumentTemplate(ctx context.Context, req *documentsv1.DocumentTemplatePatchRequest, params documentsv1.PatchDocumentTemplateParams) (documentsv1.PatchDocumentTemplateRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.PatchDocumentTemplateUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	item, err := a.documentService.UpdateTemplate(ctx, schoolID, params.TemplateId.String(), converter.ConvertTemplatePatchRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidArgument):
			res := documentsv1.PatchDocumentTemplateBadRequest(newErrorBody(ctx, "bad_request", "invalid request"))
			return &res, nil
		case errors.Is(err, repo.ErrNotFound):
			res := documentsv1.PatchDocumentTemplateNotFound(newErrorBody(ctx, "not_found", "template not found"))
			return &res, nil
		case errors.Is(err, repo.ErrConflict):
			res := documentsv1.PatchDocumentTemplateConflict(newErrorBody(ctx, "conflict", "resource conflict"))
			return &res, nil
		default:
			return nil, err
		}
	}

	data, err := converter.ConvertDocumentTemplate(item)
	if err != nil {
		return nil, err
	}
	return &documentsv1.DocumentTemplateResponse{Data: data}, nil
}

func (a *api) DeleteDocumentTemplate(ctx context.Context, params documentsv1.DeleteDocumentTemplateParams) (documentsv1.DeleteDocumentTemplateRes, error) {
	schoolID, err := a.schoolID(ctx)
	if err != nil {
		res := documentsv1.DeleteDocumentTemplateUnauthorized(newErrorBody(ctx, "unauthorized", "missing auth context"))
		return &res, nil
	}

	if err := a.documentService.DeleteTemplate(ctx, schoolID, params.TemplateId.String()); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			res := documentsv1.DeleteDocumentTemplateNotFound(newErrorBody(ctx, "not_found", "template not found"))
			return &res, nil
		}
		return nil, err
	}
	return &documentsv1.DeleteDocumentTemplateNoContent{}, nil
}
