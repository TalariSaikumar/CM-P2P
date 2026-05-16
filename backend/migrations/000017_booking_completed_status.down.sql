BEGIN;

UPDATE bookings
SET status = 'CONFIRMED'
WHERE status = 'COMPLETED';

COMMIT;
