BEGIN;

CREATE TABLE IF NOT EXISTS booking_handover_photos (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id uuid NOT NULL REFERENCES bookings(id) ON UPDATE CASCADE ON DELETE CASCADE,
    uploader_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    step varchar(32) NOT NULL,
    blob_path varchar(512) NOT NULL,
    blob_url varchar(1024) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS idx_booking_handover_photos_booking_id ON booking_handover_photos (booking_id);
CREATE INDEX IF NOT EXISTS idx_booking_handover_photos_booking_step ON booking_handover_photos (booking_id, step);

COMMIT;
