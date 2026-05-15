-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE schools (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT schools_name_not_blank
        CHECK (btrim(name) <> ''),
    CONSTRAINT schools_status_check
        CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE UNIQUE INDEX schools_name_active_udx
    ON schools (name)
    WHERE deleted_at IS NULL;

CREATE INDEX schools_status_active_idx
    ON schools (status, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE branches (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    name text NOT NULL,
    address text NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT branches_name_not_blank
        CHECK (btrim(name) <> ''),
    CONSTRAINT branches_status_check
        CHECK (status IN ('active', 'inactive')),
    CONSTRAINT branches_school_fk
        FOREIGN KEY (school_id) REFERENCES schools(id)
);

CREATE UNIQUE INDEX branches_school_name_active_udx
    ON branches (school_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX branches_school_status_active_idx
    ON branches (school_id, status, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS branches_school_status_active_idx;
DROP INDEX IF EXISTS branches_school_name_active_udx;
DROP INDEX IF EXISTS schools_status_active_idx;
DROP INDEX IF EXISTS schools_name_active_udx;

DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS schools;
