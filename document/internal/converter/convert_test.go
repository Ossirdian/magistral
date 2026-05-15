package converter

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ossirdian/magistral/document/internal/model"
	documentsv1 "github.com/ossirdian/magistral/shared/pkg/documents/v1"
)

func TestConvertUploadRequestDecodesBase64(t *testing.T) {
	t.Parallel()

	req := &documentsv1.DocumentUploadRequest{
		StudentID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Type:          "contract",
		Title:         "Agreement",
		MimeType:      "application/pdf",
		FileName:      "agreement.pdf",
		ContentBase64: base64.StdEncoding.EncodeToString([]byte("payload")),
	}

	got, err := ConvertUploadRequest(req)
	if err != nil {
		t.Fatalf("ConvertUploadRequest returned error: %v", err)
	}
	if got.StudentID != req.StudentID.String() || string(got.Content) != "payload" {
		t.Fatalf("unexpected upload conversion: %+v", got)
	}
}

func TestConvertTemplatePatchRequestPreservesNulls(t *testing.T) {
	t.Parallel()

	var name documentsv1.OptNilString
	name.SetToNull()
	req := &documentsv1.DocumentTemplatePatchRequest{
		Name:    name,
		Type:    documentsv1.NewOptNilString("contract"),
		Content: documentsv1.NewOptNilString("body"),
	}

	got := ConvertTemplatePatchRequest(req)

	if !got.NameSet || got.Name != nil {
		t.Fatalf("unexpected name patch: %+v", got)
	}
	if !got.TypeSet || got.Type == nil || *got.Type != "contract" {
		t.Fatalf("unexpected type patch: %+v", got)
	}
	if !got.ContentSet || got.Content == nil || *got.Content != "body" {
		t.Fatalf("unexpected content patch: %+v", got)
	}
}

func TestConvertDocumentAndDownload(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	document, err := ConvertDocument(model.Document{
		ID:        "22222222-2222-2222-2222-222222222222",
		SchoolID:  "33333333-3333-3333-3333-333333333333",
		StudentID: "44444444-4444-4444-4444-444444444444",
		Type:      "contract",
		Title:     "Agreement",
		Status:    model.DocumentStatusUploaded,
		ObjectKey: "schools/a/students/b/documents/c.pdf",
		MimeType:  "application/pdf",
		Size:      128,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ConvertDocument returned error: %v", err)
	}
	if document.Title != "Agreement" || document.Status != documentsv1.DocumentStatusUploaded {
		t.Fatalf("unexpected converted document: %+v", document)
	}

	download, err := ConvertDownload(model.DocumentDownload{
		DocumentID: "55555555-5555-5555-5555-555555555555",
		FileName:   "agreement.pdf",
		MimeType:   "application/pdf",
		Content:    []byte("payload"),
	})
	if err != nil {
		t.Fatalf("ConvertDownload returned error: %v", err)
	}
	if decoded, _ := base64.StdEncoding.DecodeString(download.ContentBase64); string(decoded) != "payload" {
		t.Fatalf("unexpected converted download: %+v", download)
	}
}
