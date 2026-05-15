-- +goose Up
CREATE TABLE documents (
    id uuid PRIMARY KEY,
    school_id uuid NOT NULL,
    student_id uuid NOT NULL,
    type text NOT NULL,
    title text NOT NULL,
    status text NOT NULL CHECK (status IN ('draft', 'generated', 'uploaded', 'archived')),
    object_key text NOT NULL UNIQUE,
    mime_type text NOT NULL,
    size bigint NOT NULL CHECK (size >= 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE TABLE document_templates (
    id uuid PRIMARY KEY,
    school_id uuid NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    content text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
);

CREATE INDEX idx_documents_school_student_active
    ON documents (school_id, student_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_documents_school_type_active
    ON documents (school_id, type, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_document_templates_school_type_active
    ON document_templates (school_id, type, created_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_document_templates_school_type_active;
DROP INDEX IF EXISTS idx_documents_school_type_active;
DROP INDEX IF EXISTS idx_documents_school_student_active;
DROP TABLE IF EXISTS document_templates;
DROP TABLE IF EXISTS documents;
