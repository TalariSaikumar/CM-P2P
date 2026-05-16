BEGIN;

-- Trips finished (return accepted + paid) move from CONFIRMED to COMPLETED.
UPDATE bookings
SET status = 'COMPLETED'
WHERE status = 'CONFIRMED'
  AND owner_return_accepted_at IS NOT NULL
  AND payment_status = 'PAID';

COMMIT;
