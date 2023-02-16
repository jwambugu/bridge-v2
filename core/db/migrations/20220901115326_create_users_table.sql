-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id             uuid primary key default gen_random_uuid(),
    name           varchar        NOT NULL,
    email          varchar UNIQUE NOT NULL,
    phone_number   varchar UNIQUE NOT NULL,
    password       varchar        NOT NULL,
    account_status varchar        NOT NULL,
    meta           jsonb            DEFAULT '{}',
    created_at     timestamptz      DEFAULT current_timestamp,
    updated_at     timestamptz      DEFAULT current_timestamp,
    deleted_at     timestamptz      DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_name ON users (name);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users (phone_number);
CREATE INDEX IF NOT EXISTS idx_users_meta ON users USING gin (meta);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
