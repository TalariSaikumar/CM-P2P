BEGIN;

DROP TABLE IF EXISTS booking_reviews;

ALTER TABLE bookings DROP COLUMN IF EXISTS cancellation_reason;
ALTER TABLE bookings DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE bookings DROP COLUMN IF EXISTS cancelled_by_user_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS pickup_odometer_km;
ALTER TABLE bookings DROP COLUMN IF EXISTS pickup_fuel_percent;
ALTER TABLE bookings DROP COLUMN IF EXISTS pickup_handover_notes;
ALTER TABLE bookings DROP COLUMN IF EXISTS pickup_handover_at;
ALTER TABLE bookings DROP COLUMN IF EXISTS return_odometer_km;
ALTER TABLE bookings DROP COLUMN IF EXISTS return_fuel_percent;
ALTER TABLE bookings DROP COLUMN IF EXISTS return_handover_notes;
ALTER TABLE bookings DROP COLUMN IF EXISTS return_handover_at;

COMMIT;
