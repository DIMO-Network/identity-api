-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS connections (
    "name" text NOT NULL PRIMARY KEY,
    address bytea NOT NULL UNIQUE CHECK (length(address) = 20),
    "owner" bytea NOT NULL CHECK (length("owner") = 20),
    minted_at timestamptz NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS connections;
-- +goose StatementEnd
