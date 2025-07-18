-- +goose Up
-- +goose StatementBegin
CREATE TABLE storage_nodes (
    "id" bytea PRIMARY KEY CHECK (length("id") = 32), -- This is the keccak256 of label.
    "label" text NOT NULL UNIQUE,
    address bytea NOT NULL UNIQUE CHECK (length(address) = 20),
    "owner" bytea NOT NULL CHECK (length("owner") = 20),
    uri text NOT NULL,
    minted_at timestamptz NOT NULL
);

ALTER TABLE vehicles
    ADD COLUMN storage_node_id bytea CHECK (length(storage_node_id) = 32) REFERENCES storage_nodes (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles
    DROP COLUMN storage_node_id;

DROP TABLE storage_nodes;
-- +goose StatementEnd
