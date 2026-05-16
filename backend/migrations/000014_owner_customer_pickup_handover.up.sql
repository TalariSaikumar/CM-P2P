BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS owner_pickup_odometer_km integer,
    ADD COLUMN IF NOT EXISTS owner_pickup_fuel_percent smallint,
    ADD COLUMN IF NOT EXISTS owner_pickup_handover_notes text,
    ADD COLUMN IF NOT EXISTS owner_pickup_handover_at timestamptz,
    ADD COLUMN IF NOT EXISTS customer_pickup_accepted_at timestamptz;

-- Legacy rows: treat existing pickup as both owner handover and customer acceptance.
UPDATE bookings
SET owner_pickup_odometer_km = pickup_odometer_km,
    owner_pickup_fuel_percent = pickup_fuel_percent,
    owner_pickup_handover_notes = pickup_handover_notes,
    owner_pickup_handover_at = pickup_handover_at,
    customer_pickup_accepted_at = pickup_handover_at
WHERE pickup_handover_at IS NOT NULL
  AND owner_pickup_handover_at IS NULL;

COMMIT;
