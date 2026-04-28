BEGIN;

CREATE TABLE IF NOT EXISTS bookings (
    unique_id SERIAL NOT NULL,
    unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED,
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    car_id uuid NOT NULL REFERENCES cars(id) ON UPDATE CASCADE ON DELETE CASCADE,
    customer_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    owner_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    status varchar(24) NOT NULL DEFAULT 'PENDING',
    final_booking_price numeric(14,2),
    customer_note text,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    deleted_at timestamptz
);

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS unique_id SERIAL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS car_id uuid;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS customer_id uuid;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS owner_id uuid;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS status varchar(24) DEFAULT 'PENDING';
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS final_booking_price numeric(14,2);
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS customer_note text;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_unique_id ON bookings (unique_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_unique_number ON bookings (unique_number);
CREATE INDEX IF NOT EXISTS idx_bookings_car_id ON bookings (car_id);
CREATE INDEX IF NOT EXISTS idx_bookings_customer_id ON bookings (customer_id);
CREATE INDEX IF NOT EXISTS idx_bookings_owner_id ON bookings (owner_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings (status);
CREATE INDEX IF NOT EXISTS idx_bookings_deleted_at ON bookings (deleted_at);

DROP TRIGGER IF EXISTS trg_bookings_touch_timestamps ON bookings;
CREATE TRIGGER trg_bookings_touch_timestamps
    BEFORE INSERT OR UPDATE ON bookings
    FOR EACH ROW
    EXECUTE PROCEDURE touch_row_timestamps_utc();

COMMIT;
