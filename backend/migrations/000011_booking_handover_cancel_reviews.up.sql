BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS cancellation_reason text,
    ADD COLUMN IF NOT EXISTS cancelled_at timestamptz,
    ADD COLUMN IF NOT EXISTS cancelled_by_user_id uuid REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS pickup_odometer_km integer,
    ADD COLUMN IF NOT EXISTS pickup_fuel_percent smallint,
    ADD COLUMN IF NOT EXISTS pickup_handover_notes text,
    ADD COLUMN IF NOT EXISTS pickup_handover_at timestamptz,
    ADD COLUMN IF NOT EXISTS return_odometer_km integer,
    ADD COLUMN IF NOT EXISTS return_fuel_percent smallint,
    ADD COLUMN IF NOT EXISTS return_handover_notes text,
    ADD COLUMN IF NOT EXISTS return_handover_at timestamptz;

CREATE TABLE IF NOT EXISTS booking_reviews (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id uuid NOT NULL REFERENCES bookings(id) ON UPDATE CASCADE ON DELETE CASCADE,
    reviewer_party varchar(16) NOT NULL,
    reviewer_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    rating smallint NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT booking_reviews_party_chk CHECK (reviewer_party IN ('CUSTOMER', 'OWNER')),
    CONSTRAINT booking_reviews_booking_party_uk UNIQUE (booking_id, reviewer_party)
);

CREATE INDEX IF NOT EXISTS idx_booking_reviews_booking_id ON booking_reviews (booking_id);

COMMIT;
