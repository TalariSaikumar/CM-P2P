BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS payment_status varchar(16) NOT NULL DEFAULT 'UNPAID',
    ADD COLUMN IF NOT EXISTS payment_method varchar(24),
    ADD COLUMN IF NOT EXISTS paid_at timestamptz,
    ADD COLUMN IF NOT EXISTS customer_commission_rate numeric(6,3),
    ADD COLUMN IF NOT EXISTS owner_commission_rate numeric(6,3),
    ADD COLUMN IF NOT EXISTS customer_commission_amount numeric(14,2),
    ADD COLUMN IF NOT EXISTS owner_commission_amount numeric(14,2),
    ADD COLUMN IF NOT EXISTS customer_total_paid numeric(14,2),
    ADD COLUMN IF NOT EXISTS owner_net_payout numeric(14,2);

CREATE INDEX IF NOT EXISTS idx_bookings_payment_status ON bookings (payment_status);

COMMIT;
