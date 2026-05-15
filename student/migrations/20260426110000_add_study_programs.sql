-- +goose Up
CREATE TABLE study_programs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    name text NOT NULL,
    transmission text NOT NULL,
    total_hours numeric(10,2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT study_programs_name_not_blank
        CHECK (btrim(name) <> ''),
    CONSTRAINT study_programs_transmission_check
        CHECK (transmission IN ('manual', 'automatic')),
    CONSTRAINT study_programs_total_hours_check
        CHECK (total_hours >= 0)
);

CREATE UNIQUE INDEX study_programs_school_name_active_uidx
    ON study_programs (school_id, lower(name))
    WHERE deleted_at IS NULL;

ALTER TABLE students
    ADD COLUMN program_id uuid NULL,
    ADD COLUMN manual_hours_completed numeric(10,2) NOT NULL DEFAULT 0;

ALTER TABLE students
    ADD CONSTRAINT students_manual_hours_completed_check
        CHECK (manual_hours_completed >= 0),
    ADD CONSTRAINT students_program_fk
        FOREIGN KEY (program_id) REFERENCES study_programs (id);

CREATE INDEX students_school_program_active_idx
    ON students (school_id, program_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS students_school_program_active_idx;
ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_program_fk;
ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_manual_hours_completed_check;
ALTER TABLE students
    DROP COLUMN IF EXISTS manual_hours_completed;
ALTER TABLE students
    DROP COLUMN IF EXISTS program_id;
DROP INDEX IF EXISTS study_programs_school_name_active_uidx;
DROP TABLE IF EXISTS study_programs;
