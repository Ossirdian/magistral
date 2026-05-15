-- +goose Up
ALTER TABLE cars
    ADD COLUMN transmission text NOT NULL DEFAULT 'manual';

ALTER TABLE cars
    ADD CONSTRAINT cars_transmission_check
    CHECK (transmission IN ('manual', 'automatic'));

-- +goose Down
ALTER TABLE cars
    DROP CONSTRAINT IF EXISTS cars_transmission_check;

ALTER TABLE cars
    DROP COLUMN IF EXISTS transmission;
