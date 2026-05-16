BEGIN;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS owner_return_accepted_at timestamptz;

-- Legacy: treat existing return handover as customer return; owner accepted when fully paid.
UPDATE bookings
SET owner_return_accepted_at = return_handover_at
WHERE return_handover_at IS NOT NULL
  AND payment_status = 'PAID'
  AND owner_return_accepted_at IS NULL;

COMMIT;
