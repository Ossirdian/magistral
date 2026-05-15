package model

import "time"

type Document struct {
	ID        string
	SchoolID  string
	StudentID string
	Type      string
	Title     string
	Status    DocumentStatus
	ObjectKey string
	MimeType  string
	Size      int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentUpload struct {
	StudentID string
	Type      string
	Title     string
	MimeType  string
	FileName  string
	Content   []byte
}

type DocumentGenerate struct {
	StudentID    string
	TemplateID   string
	Title        string
	FileName     *string
	Placeholders map[string]string
}

type DocumentDownload struct {
	DocumentID string
	FileName   string
	MimeType   string
	Content    []byte
}

type DocumentFilter struct {
	StudentID *string
	Type      *string
	Status    *DocumentStatus
	Limit     int
	Offset    int
}

type DocumentTemplate struct {
	ID        string
	SchoolID  string
	Name      string
	Type      string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentTemplateCreate struct {
	Name    string
	Type    string
	Content string
}

type DocumentTemplatePatch struct {
	Name       *string
	NameSet    bool
	Type       *string
	TypeSet    bool
	Content    *string
	ContentSet bool
}

type DocumentTemplateFilter struct {
	Type   *string
	Limit  int
	Offset int
}

type PageMeta struct {
	Limit  int
	Offset int
	Count  int
}

type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusGenerated DocumentStatus = "generated"
	DocumentStatusUploaded  DocumentStatus = "uploaded"
	DocumentStatusArchived  DocumentStatus = "archived"
)
