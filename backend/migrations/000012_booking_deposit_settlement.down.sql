BEGIN;

DROP INDEX IF EXISTS idx_booking_post_trip_charges_booking_id;
DROP TABLE IF EXISTS booking_post_trip_charges;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS customer_acknowledged_terms_at,
    DROP COLUMN IF EXISTS settlement_submitted_at,
    DROP COLUMN IF EXISTS post_trip_charges_total,
    DROP COLUMN IF EXISTS deposit_customer_total,
    DROP COLUMN IF EXISTS deposit_paid_at;

COMMIT;
