BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS deposit_paid_at timestamptz,
    ADD COLUMN IF NOT EXISTS deposit_customer_total numeric(14,2),
    ADD COLUMN IF NOT EXISTS post_trip_charges_total numeric(14,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS settlement_submitted_at timestamptz,
    ADD COLUMN IF NOT EXISTS customer_acknowledged_terms_at timestamptz;

UPDATE bookings SET post_trip_charges_total = 0 WHERE post_trip_charges_total IS NULL;

CREATE TABLE IF NOT EXISTS booking_post_trip_charges (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id uuid        NOT NULL REFERENCES bookings (id) ON UPDATE CASCADE ON DELETE CASCADE,
    label      text        NOT NULL,
    amount_inr numeric(14, 2) NOT NULL CHECK (amount_inr >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_booking_post_trip_charges_booking_id ON booking_post_trip_charges (booking_id);

COMMIT;
