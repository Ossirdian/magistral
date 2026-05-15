-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE students (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    full_name text NOT NULL,
    phone text NOT NULL,
    status text NOT NULL DEFAULT 'new',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT students_full_name_not_blank
        CHECK (btrim(full_name) <> ''),
    CONSTRAINT students_phone_not_blank
        CHECK (btrim(phone) <> ''),
    CONSTRAINT students_status_check
        CHECK (status IN ('new', 'active', 'paused', 'graduated', 'archived'))
);

CREATE INDEX students_school_branch_active_idx
    ON students (school_id, branch_id)
    WHERE deleted_at IS NULL;

CREATE INDEX students_school_status_active_idx
    ON students (school_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX students_school_name_active_idx
    ON students (school_id, full_name)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS students_school_name_active_idx;
DROP INDEX IF EXISTS students_school_status_active_idx;
DROP INDEX IF EXISTS students_school_branch_active_idx;
DROP TABLE IF EXISTS students;
