BEGIN;

DROP INDEX IF EXISTS idx_bookings_payment_status;

ALTER TABLE bookings DROP COLUMN IF EXISTS payment_status;
ALTER TABLE bookings DROP COLUMN IF EXISTS payment_method;
ALTER TABLE bookings DROP COLUMN IF EXISTS paid_at;
ALTER TABLE bookings DROP COLUMN IF EXISTS customer_commission_rate;
ALTER TABLE bookings DROP COLUMN IF EXISTS owner_commission_rate;
ALTER TABLE bookings DROP COLUMN IF EXISTS customer_commission_amount;
ALTER TABLE bookings DROP COLUMN IF EXISTS owner_commission_amount;
ALTER TABLE bookings DROP COLUMN IF EXISTS customer_total_paid;
ALTER TABLE bookings DROP COLUMN IF EXISTS owner_net_payout;

COMMIT;
