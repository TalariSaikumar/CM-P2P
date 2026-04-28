BEGIN;

CREATE TABLE IF NOT EXISTS cars (
    unique_id SERIAL NOT NULL,
    unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED,
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    car_name varchar(255) NOT NULL,
    car_model varchar(255) NOT NULL,
    car_number varchar(32) NOT NULL UNIQUE,
    registration_number varchar(64) NOT NULL,
    engine_number varchar(64) NOT NULL,
    price_per_hour numeric(12,2) NOT NULL,
    price_per_day numeric(12,2) NOT NULL,
    price_per_km numeric(12,2) NOT NULL,
    location varchar(255) NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    deleted_at timestamptz
);

ALTER TABLE cars ADD COLUMN IF NOT EXISTS unique_id SERIAL;
ALTER TABLE cars ADD COLUMN IF NOT EXISTS unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED;
ALTER TABLE cars ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE cars ADD COLUMN IF NOT EXISTS owner_id uuid;
ALTER TABLE cars ADD COLUMN IF NOT EXISTS car_name varchar(255);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS car_model varchar(255);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS car_number varchar(32);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS registration_number varchar(64);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS engine_number varchar(64);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS price_per_hour numeric(12,2);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS price_per_day numeric(12,2);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS price_per_km numeric(12,2);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS location varchar(255);
ALTER TABLE cars ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true;
ALTER TABLE cars ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE cars ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE cars ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cars_unique_id ON cars (unique_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cars_unique_number ON cars (unique_number);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cars_car_number ON cars (car_number);
CREATE INDEX IF NOT EXISTS idx_cars_owner_id ON cars (owner_id);
CREATE INDEX IF NOT EXISTS idx_cars_car_model ON cars (car_model);
CREATE INDEX IF NOT EXISTS idx_cars_location ON cars (location);
CREATE INDEX IF NOT EXISTS idx_cars_is_active ON cars (is_active);
CREATE INDEX IF NOT EXISTS idx_cars_deleted_at ON cars (deleted_at);

DROP TRIGGER IF EXISTS trg_cars_touch_timestamps ON cars;
CREATE TRIGGER trg_cars_touch_timestamps
    BEFORE INSERT OR UPDATE ON cars
    FOR EACH ROW
    EXECUTE PROCEDURE touch_row_timestamps_utc();

COMMIT;
