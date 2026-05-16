BEGIN;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS customer_accepted_price_amount,
    DROP COLUMN IF EXISTS customer_accepted_price_at;

COMMIT;
