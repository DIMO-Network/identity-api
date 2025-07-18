-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_nodes (
    "id" bytea PRIMARY KEY CHECK (length("id") = 32), -- This is the keccak256 of label.
    "label" text NOT NULL UNIQUE,
    address bytea NOT NULL UNIQUE CHECK (length(address) = 20),
    "owner" bytea NOT NULL CHECK (length("owner") = 20),
    uri TEXT NOT NULL,
    minted_at timestamptz NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS storage_nodes;
-- +goose StatementEnd
