-- +goose Up
-- +goose StatementBegin
CREATE TABLE dcns (
    node bytea CONSTRAINT dcns_node_check CHECK (length(node) = 32) CONSTRAINT dcns_pkey PRIMARY KEY,
    owner_address bytea CONSTRAINT dcns_owner_address_check CHECK (length(owner_address) = 20) NOT NULL,
    expiration timestamptz(0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE dcns;
-- +goose StatementEnd
