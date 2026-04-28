BEGIN;

CREATE TABLE IF NOT EXISTS car_images (
    unique_id SERIAL NOT NULL,
    unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED,
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    car_id uuid NOT NULL REFERENCES cars(id) ON UPDATE CASCADE ON DELETE CASCADE,
    blob_path varchar(512) NOT NULL,
    blob_url varchar(1024) NOT NULL,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    deleted_at timestamptz
);

ALTER TABLE car_images ADD COLUMN IF NOT EXISTS unique_id SERIAL;
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED;
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS car_id uuid;
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS blob_path varchar(512);
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS blob_url varchar(1024);
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS sort_order integer DEFAULT 0;
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE car_images ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS idx_car_images_unique_id ON car_images (unique_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_car_images_unique_number ON car_images (unique_number);
CREATE INDEX IF NOT EXISTS idx_car_images_car_id ON car_images (car_id);
CREATE INDEX IF NOT EXISTS idx_car_images_deleted_at ON car_images (deleted_at);

DROP TRIGGER IF EXISTS trg_car_images_touch_timestamps ON car_images;
CREATE TRIGGER trg_car_images_touch_timestamps
    BEFORE INSERT OR UPDATE ON car_images
    FOR EACH ROW
    EXECUTE PROCEDURE touch_row_timestamps_utc();

COMMIT;
