-- +goose Up
-- +goose StatementBegin
CREATE TABLE dcn (
    node bytea CONSTRAINT dcn_node_check CHECK (length(node) = 32) PRIMARY KEY,
    owner_address bytea CONSTRAINT dcn_owner_address_check CHECK (length(owner_address) = 20) NOT NULL,
    resolver_address bytea CONSTRAINT dcn_resolver_address_check CHECK (length(resolver_address) = 20),
    expiration timestamptz(0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE dcn;
-- +goose StatementEnd
