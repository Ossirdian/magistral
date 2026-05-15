-- +goose Up
CREATE TABLE groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT groups_name_not_blank
        CHECK (btrim(name) <> '')
);

CREATE INDEX groups_school_branch_active_idx
    ON groups (school_id, branch_id)
    WHERE deleted_at IS NULL;

CREATE INDEX groups_school_name_active_idx
    ON groups (school_id, name)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS groups_school_name_active_idx;
DROP INDEX IF EXISTS groups_school_branch_active_idx;
DROP TABLE IF EXISTS groups;
