BEGIN;

ALTER TABLE bookings DROP COLUMN IF EXISTS gst_percent_on_commission;
ALTER TABLE bookings DROP COLUMN IF EXISTS customer_gst_amount;
ALTER TABLE bookings DROP COLUMN IF EXISTS owner_gst_amount;

COMMIT;
