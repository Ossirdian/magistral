package converter

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/document/internal/model"
	svc "github.com/ossirdian/magistral/document/internal/service"
	documentsv1 "github.com/ossirdian/magistral/shared/pkg/documents/v1"
)

func ConvertDocument(item model.Document) (documentsv1.Document, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return documentsv1.Document{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return documentsv1.Document{}, err
	}
	studentID, err := parseUUID(item.StudentID)
	if err != nil {
		return documentsv1.Document{}, err
	}
	return documentsv1.Document{
		ID:        id,
		SchoolID:  schoolID,
		StudentID: studentID,
		Type:      item.Type,
		Title:     item.Title,
		Status:    documentsv1.DocumentStatus(item.Status),
		ObjectKey: item.ObjectKey,
		MimeType:  item.MimeType,
		Size:      item.Size,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

func ConvertDocuments(items []model.Document) ([]documentsv1.Document, error) {
	result := make([]documentsv1.Document, 0, len(items))
	for _, item := range items {
		converted, err := ConvertDocument(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertDocumentTemplate(item model.DocumentTemplate) (documentsv1.DocumentTemplate, error) {
	id, err := parseUUID(item.ID)
	if err != nil {
		return documentsv1.DocumentTemplate{}, err
	}
	schoolID, err := parseUUID(item.SchoolID)
	if err != nil {
		return documentsv1.DocumentTemplate{}, err
	}
	return documentsv1.DocumentTemplate{
		ID:        id,
		SchoolID:  schoolID,
		Name:      item.Name,
		Type:      item.Type,
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

func ConvertDocumentTemplates(items []model.DocumentTemplate) ([]documentsv1.DocumentTemplate, error) {
	result := make([]documentsv1.DocumentTemplate, 0, len(items))
	for _, item := range items {
		converted, err := ConvertDocumentTemplate(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ConvertDownload(item model.DocumentDownload) (documentsv1.DocumentDownload, error) {
	documentID, err := parseUUID(item.DocumentID)
	if err != nil {
		return documentsv1.DocumentDownload{}, err
	}
	return documentsv1.DocumentDownload{
		DocumentID:    documentID,
		FileName:      item.FileName,
		MimeType:      item.MimeType,
		ContentBase64: base64.StdEncoding.EncodeToString(item.Content),
	}, nil
}

func ConvertUploadRequest(req *documentsv1.DocumentUploadRequest) (model.DocumentUpload, error) {
	content, err := base64.StdEncoding.DecodeString(req.ContentBase64)
	if err != nil {
		return model.DocumentUpload{}, svc.ErrInvalidArgument
	}
	return model.DocumentUpload{
		StudentID: req.StudentID.String(),
		Type:      req.Type,
		Title:     req.Title,
		MimeType:  req.MimeType,
		FileName:  req.FileName,
		Content:   content,
	}, nil
}

func ConvertGenerateRequest(req *documentsv1.DocumentGenerateRequest) model.DocumentGenerate {
	placeholders := map[string]string{}
	if value, ok := req.Placeholders.Get(); ok {
		placeholders = map[string]string(value)
	}
	var fileName *string
	if value, ok := req.FileName.Get(); ok {
		fileName = &value
	}
	return model.DocumentGenerate{
		StudentID:    req.StudentID.String(),
		TemplateID:   req.TemplateID.String(),
		Title:        req.Title,
		FileName:     fileName,
		Placeholders: placeholders,
	}
}

func ConvertDocumentFilter(params documentsv1.ListDocumentsParams) model.DocumentFilter {
	return model.DocumentFilter{
		StudentID: optUUIDToPtr(params.StudentID),
		Type:      optStringToPtr(params.Type),
		Status:    optDocumentStatusToPtr(params.Status),
		Limit:     params.Limit.Or(0),
		Offset:    params.Offset.Or(0),
	}
}

func ConvertTemplateFilter(params documentsv1.ListDocumentTemplatesParams) model.DocumentTemplateFilter {
	return model.DocumentTemplateFilter{
		Type:   optStringToPtr(params.Type),
		Limit:  params.Limit.Or(0),
		Offset: params.Offset.Or(0),
	}
}

func ConvertTemplateCreateRequest(req *documentsv1.DocumentTemplateCreateRequest) model.DocumentTemplateCreate {
	return model.DocumentTemplateCreate{
		Name:    req.Name,
		Type:    req.Type,
		Content: req.Content,
	}
}

func ConvertTemplatePatchRequest(req *documentsv1.DocumentTemplatePatchRequest) model.DocumentTemplatePatch {
	name, nameSet := optNilStringToPatch(req.Name)
	typ, typeSet := optNilStringToPatch(req.Type)
	content, contentSet := optNilStringToPatch(req.Content)
	return model.DocumentTemplatePatch{
		Name:       name,
		NameSet:    nameSet,
		Type:       typ,
		TypeSet:    typeSet,
		Content:    content,
		ContentSet: contentSet,
	}
}

func ConvertMeta(meta model.PageMeta) documentsv1.Meta {
	return documentsv1.Meta{Limit: meta.Limit, Offset: meta.Offset, Count: meta.Count}
}

func optUUIDToPtr(v documentsv1.OptUUID) *string {
	if value, ok := v.Get(); ok {
		s := value.String()
		return &s
	}
	return nil
}

func optStringToPtr(v documentsv1.OptString) *string {
	if value, ok := v.Get(); ok {
		return &value
	}
	return nil
}

func optDocumentStatusToPtr(v documentsv1.OptDocumentStatus) *model.DocumentStatus {
	if value, ok := v.Get(); ok {
		converted := model.DocumentStatus(value)
		return &converted
	}
	return nil
}

func optNilStringToPatch(v documentsv1.OptNilString) (*string, bool) {
	if !v.IsSet() {
		return nil, false
	}
	if v.IsNull() {
		return nil, true
	}
	value, ok := v.Get()
	if !ok {
		return nil, true
	}
	return &value, true
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid %q: %w", value, err)
	}
	return id, nil
}
