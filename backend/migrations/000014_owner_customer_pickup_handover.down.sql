BEGIN;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS customer_pickup_accepted_at,
    DROP COLUMN IF EXISTS owner_pickup_handover_at,
    DROP COLUMN IF EXISTS owner_pickup_handover_notes,
    DROP COLUMN IF EXISTS owner_pickup_fuel_percent,
    DROP COLUMN IF EXISTS owner_pickup_odometer_km;

COMMIT;
