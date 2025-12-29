BEGIN;

ALTER TABLE users
    ADD COLUMN password_hash TEXT NOT NULL,
    ADD COLUMN is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN created_at    timestamptz NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at    timestamptz NOT NULL DEFAULT NOW();


CREATE INDEX IF NOT EXISTS users_login_idx ON users (login);
CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
