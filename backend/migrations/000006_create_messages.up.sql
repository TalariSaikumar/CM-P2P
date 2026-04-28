BEGIN;

CREATE TABLE IF NOT EXISTS messages (
    unique_id SERIAL NOT NULL,
    unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED,
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id uuid NOT NULL REFERENCES bookings(id) ON UPDATE CASCADE ON DELETE CASCADE,
    sender_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    deleted_at timestamptz
);

ALTER TABLE messages ADD COLUMN IF NOT EXISTS unique_id SERIAL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE messages ADD COLUMN IF NOT EXISTS booking_id uuid;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS sender_id uuid;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS body text;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE messages ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE messages ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_unique_id ON messages (unique_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_unique_number ON messages (unique_number);
CREATE INDEX IF NOT EXISTS idx_messages_booking_id ON messages (booking_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages (deleted_at);

DROP TRIGGER IF EXISTS trg_messages_touch_timestamps ON messages;
CREATE TRIGGER trg_messages_touch_timestamps
    BEFORE INSERT OR UPDATE ON messages
    FOR EACH ROW
    EXECUTE PROCEDURE touch_row_timestamps_utc();

COMMIT;
