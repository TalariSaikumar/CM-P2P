BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS gst_percent_on_commission numeric(5,2),
    ADD COLUMN IF NOT EXISTS customer_gst_amount numeric(14,2),
    ADD COLUMN IF NOT EXISTS owner_gst_amount numeric(14,2);

COMMIT;
