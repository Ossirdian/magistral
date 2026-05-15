-- +goose Up
CREATE TABLE lessons (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    branch_id uuid NULL,
    student_id uuid NOT NULL,
    instructor_id uuid NOT NULL,
    car_id uuid NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'scheduled',
    comment text NULL,
    is_generated boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE INDEX lessons_school_start_idx
    ON lessons (school_id, start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX lessons_school_branch_start_idx
    ON lessons (school_id, branch_id, start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX lessons_school_student_start_idx
    ON lessons (school_id, student_id, start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX lessons_school_instructor_start_idx
    ON lessons (school_id, instructor_id, start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX lessons_school_car_start_idx
    ON lessons (school_id, car_id, start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX lessons_school_generated_idx
    ON lessons (school_id, is_generated, start_at)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE lessons;
