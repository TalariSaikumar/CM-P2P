BEGIN;

CREATE TABLE IF NOT EXISTS kyc_attachments (
    unique_id SERIAL NOT NULL,
    unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED,
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    kind varchar(32) NOT NULL,
    blob_path varchar(512) NOT NULL,
    blob_url varchar(1024) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    deleted_at timestamptz
);

ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS unique_id SERIAL;
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED;
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS user_id uuid;
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS kind varchar(32);
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS blob_path varchar(512);
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS blob_url varchar(1024);
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE kyc_attachments ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS idx_kyc_attachments_unique_id ON kyc_attachments (unique_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_kyc_attachments_unique_number ON kyc_attachments (unique_number);
CREATE INDEX IF NOT EXISTS idx_kyc_attachments_user_id ON kyc_attachments (user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_attachments_deleted_at ON kyc_attachments (deleted_at);

DROP TRIGGER IF EXISTS trg_kyc_attachments_touch_timestamps ON kyc_attachments;
CREATE TRIGGER trg_kyc_attachments_touch_timestamps
    BEFORE INSERT OR UPDATE ON kyc_attachments
    FOR EACH ROW
    EXECUTE PROCEDURE touch_row_timestamps_utc();

COMMIT;
