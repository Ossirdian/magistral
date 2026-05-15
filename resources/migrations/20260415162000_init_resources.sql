-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE instructors (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    full_name text NOT NULL,
    phone text NULL,
    email text NULL,
    status text NOT NULL DEFAULT 'active',
    availability_status text NOT NULL DEFAULT 'available',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT instructors_full_name_not_blank
        CHECK (btrim(full_name) <> ''),
    CONSTRAINT instructors_status_check
        CHECK (status IN ('active', 'inactive', 'on_leave')),
    CONSTRAINT instructors_availability_status_check
        CHECK (availability_status IN ('available', 'unavailable', 'partial'))
);

CREATE UNIQUE INDEX instructors_id_school_uidx
    ON instructors (id, school_id);

CREATE INDEX instructors_school_branch_active_idx
    ON instructors (school_id, branch_id)
    WHERE deleted_at IS NULL;

CREATE INDEX instructors_school_status_active_idx
    ON instructors (school_id, status, availability_status)
    WHERE deleted_at IS NULL;

CREATE TABLE cars (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    instructor_id uuid NULL,
    plate_number text NOT NULL,
    brand text NULL,
    model text NULL,
    color text NULL,
    status text NOT NULL DEFAULT 'active',
    availability_status text NOT NULL DEFAULT 'available',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT cars_plate_number_not_blank
        CHECK (btrim(plate_number) <> ''),
    CONSTRAINT cars_status_check
        CHECK (status IN ('active', 'inactive', 'maintenance')),
    CONSTRAINT cars_availability_status_check
        CHECK (availability_status IN ('available', 'unavailable', 'partial')),
    CONSTRAINT cars_instructor_school_fk
        FOREIGN KEY (instructor_id, school_id)
        REFERENCES instructors (id, school_id)
);

CREATE UNIQUE INDEX cars_school_plate_number_active_udx
    ON cars (school_id, plate_number)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX cars_school_instructor_active_udx
    ON cars (school_id, instructor_id)
    WHERE deleted_at IS NULL AND instructor_id IS NOT NULL;

CREATE INDEX cars_school_branch_active_idx
    ON cars (school_id, branch_id)
    WHERE deleted_at IS NULL;

CREATE INDEX cars_school_status_active_idx
    ON cars (school_id, status, availability_status)
    WHERE deleted_at IS NULL;

CREATE TABLE instructor_availability (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id uuid NOT NULL,
    type text NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    comment text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT instructor_availability_type_check
        CHECK (type IN ('working_hours', 'vacation', 'blackout')),
    CONSTRAINT instructor_availability_time_range_check
        CHECK (start_at < end_at),
    CONSTRAINT instructor_availability_instructor_fk
        FOREIGN KEY (instructor_id) REFERENCES instructors(id)
        ON DELETE CASCADE
);

CREATE INDEX instructor_availability_lookup_idx
    ON instructor_availability (instructor_id, start_at, end_at);

CREATE TABLE car_availability (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    car_id uuid NOT NULL,
    type text NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    comment text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT car_availability_type_check
        CHECK (type IN ('working_hours', 'vacation', 'blackout')),
    CONSTRAINT car_availability_time_range_check
        CHECK (start_at < end_at),
    CONSTRAINT car_availability_car_fk
        FOREIGN KEY (car_id) REFERENCES cars(id)
        ON DELETE CASCADE
);

CREATE INDEX car_availability_lookup_idx
    ON car_availability (car_id, start_at, end_at);

-- +goose Down
DROP INDEX IF EXISTS car_availability_lookup_idx;
DROP INDEX IF EXISTS instructor_availability_lookup_idx;
DROP INDEX IF EXISTS cars_school_status_active_idx;
DROP INDEX IF EXISTS cars_school_branch_active_idx;
DROP INDEX IF EXISTS cars_school_instructor_active_udx;
DROP INDEX IF EXISTS cars_school_plate_number_active_udx;
DROP INDEX IF EXISTS instructors_school_status_active_idx;
DROP INDEX IF EXISTS instructors_school_branch_active_idx;
DROP INDEX IF EXISTS instructors_id_school_uidx;

DROP TABLE IF EXISTS car_availability;
DROP TABLE IF EXISTS instructor_availability;
DROP TABLE IF EXISTS cars;
DROP TABLE IF EXISTS instructors;