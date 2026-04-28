BEGIN;

CREATE TABLE IF NOT EXISTS users (
    unique_id SERIAL NOT NULL,
    unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED,
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    role varchar(20) NOT NULL,
    full_name varchar(255) NOT NULL,
    aadhaar_number varchar(32) NOT NULL,
    phone_number varchar(32) NOT NULL,
    address text NOT NULL,
    driving_license_number varchar(64),
    is_kyc_verified boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    deleted_at timestamptz
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS unique_id SERIAL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS unique_number INTEGER GENERATED ALWAYS AS (10000000 + unique_id - 1) STORED;
ALTER TABLE users ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE users ADD COLUMN IF NOT EXISTS email varchar(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash varchar(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS role varchar(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name varchar(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS aadhaar_number varchar(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number varchar(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS address text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS driving_license_number varchar(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_kyc_verified boolean DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at timestamptz DEFAULT clock_timestamp();
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_unique_id ON users (unique_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_unique_number ON users (unique_number);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users (phone_number);
CREATE INDEX IF NOT EXISTS idx_users_is_kyc_verified ON users (is_kyc_verified);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

DROP TRIGGER IF EXISTS trg_users_touch_timestamps ON users;
CREATE TRIGGER trg_users_touch_timestamps
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE touch_row_timestamps_utc();

COMMIT;
