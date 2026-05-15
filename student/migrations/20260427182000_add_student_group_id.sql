-- +goose Up
ALTER TABLE students
    ADD COLUMN group_id uuid NULL;

CREATE UNIQUE INDEX groups_id_school_uidx
    ON groups (id, school_id);

ALTER TABLE students
    ADD CONSTRAINT students_group_fk
    FOREIGN KEY (group_id, school_id) REFERENCES groups (id, school_id);

CREATE INDEX students_school_group_active_idx
    ON students (school_id, group_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS students_school_group_active_idx;

ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_group_fk;

DROP INDEX IF EXISTS groups_id_school_uidx;

ALTER TABLE students
    DROP COLUMN IF EXISTS group_id;
