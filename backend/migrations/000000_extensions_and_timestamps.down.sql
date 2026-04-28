BEGIN;

DROP FUNCTION IF EXISTS touch_row_timestamps_utc() CASCADE;

-- Optional: remove extension only if nothing else in the DB needs it.
-- DROP EXTENSION IF EXISTS pgcrypto;

COMMIT;
