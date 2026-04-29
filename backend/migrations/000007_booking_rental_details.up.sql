BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS rental_from timestamptz NOT NULL DEFAULT (timezone('utc', now())),
    ADD COLUMN IF NOT EXISTS rental_to timestamptz NOT NULL DEFAULT (timezone('utc', now()) + interval '1 day'),
    ADD COLUMN IF NOT EXISTS pickup_point text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS drop_point text NOT NULL DEFAULT '';

COMMIT;
