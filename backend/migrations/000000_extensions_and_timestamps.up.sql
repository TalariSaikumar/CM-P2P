BEGIN;

-- gen_random_uuid() is built-in on PostgreSQL 15+; extension keeps older versions working.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Sets created_at + updated_at on INSERT, and updated_at on UPDATE (current instant, stored as timestamptz).
CREATE OR REPLACE FUNCTION touch_row_timestamps_utc()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    ts timestamptz := clock_timestamp();
BEGIN
    IF TG_OP = 'INSERT' THEN
        NEW.created_at := ts;
        NEW.updated_at := ts;
    ELSIF TG_OP = 'UPDATE' THEN
        NEW.updated_at := ts;
    END IF;
    RETURN NEW;
END;
$$;

COMMIT;
