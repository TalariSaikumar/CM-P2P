BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS customer_accepted_price_at timestamptz,
    ADD COLUMN IF NOT EXISTS customer_accepted_price_amount numeric(14,2);

COMMIT;
