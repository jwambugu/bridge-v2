-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS branches
(
    id         bigserial primary key not null,
    name       varchar               NOT NULL,
    sap_id     smallserial           not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

CREATE INDEX IF NOT EXISTS idx_branches_name ON branches (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS branches;
-- +goose StatementEnd
