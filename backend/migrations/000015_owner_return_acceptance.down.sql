BEGIN;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS owner_return_accepted_at;

COMMIT;
