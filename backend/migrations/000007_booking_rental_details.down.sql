BEGIN;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS rental_from,
    DROP COLUMN IF EXISTS rental_to,
    DROP COLUMN IF EXISTS pickup_point,
    DROP COLUMN IF EXISTS drop_point;

COMMIT;
