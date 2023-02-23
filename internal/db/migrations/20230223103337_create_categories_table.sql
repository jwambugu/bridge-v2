-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS categories
(
    id         uuid primary key default gen_random_uuid(),
    name       varchar UNIQUE NOT NULL,
    slug       varchar UNIQUE NOT NULL,
    status     varchar        NOT NULL,
    meta       jsonb            DEFAULT '{}'::jsonb,
    created_at timestamptz      DEFAULT current_timestamp,
    updated_at timestamptz      DEFAULT current_timestamp,
    deleted_at timestamptz      DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_categories_meta ON categories USING gin (meta);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
