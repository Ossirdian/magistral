-- +goose Up
ALTER TABLE students
    ADD COLUMN preferred_weekdays text[] NOT NULL DEFAULT '{}'::text[],
    ADD COLUMN preferred_time_windows text[] NOT NULL DEFAULT '{}'::text[];

ALTER TABLE students
    ADD CONSTRAINT students_preferred_weekdays_check
        CHECK (preferred_weekdays <@ ARRAY['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday']::text[]),
    ADD CONSTRAINT students_preferred_time_windows_check
        CHECK (preferred_time_windows <@ ARRAY['morning', 'daytime', 'evening']::text[]);

-- +goose Down
ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_preferred_time_windows_check;
ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_preferred_weekdays_check;
ALTER TABLE students
    DROP COLUMN IF EXISTS preferred_time_windows;
ALTER TABLE students
    DROP COLUMN IF EXISTS preferred_weekdays;
