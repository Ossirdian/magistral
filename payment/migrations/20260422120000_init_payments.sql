-- +goose Up
CREATE TABLE payments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id uuid NOT NULL,
    student_id uuid NOT NULL,
    type text NOT NULL,
    amount numeric(12,2) NOT NULL,
    comment text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CONSTRAINT payments_amount_check CHECK (amount > 0),
    CONSTRAINT payments_type_check CHECK (type IN ('cash', 'card', 'transfer', 'other'))
);

CREATE TABLE student_contracts (
    student_id uuid PRIMARY KEY,
    school_id uuid NOT NULL,
    contract_total numeric(12,2) NOT NULL,
    remaining_amount numeric(12,2) NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX payments_school_student_active_idx
    ON payments (school_id, student_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX payments_school_type_active_idx
    ON payments (school_id, type, created_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS payments_school_type_active_idx;
DROP INDEX IF EXISTS payments_school_student_active_idx;
DROP TABLE IF EXISTS student_contracts;
DROP TABLE IF EXISTS payments;
