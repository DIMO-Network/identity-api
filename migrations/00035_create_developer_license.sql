-- +goose Up
-- +goose StatementBegin
CREATE TABLE developer_licenses(
    token_id int NOT NULL CONSTRAINT developer_licenses_pkey PRIMARY KEY,
    owner bytea NOT NULL CONSTRAINT developer_licenses_owner_check CHECK (length(owner) = 20),
    address bytea NOT NULL CONSTRAINT developer_licenses_address_check CHECK (length(address) = 20) CONSTRAINT developer_licenses_address_key UNIQUE,
    minted_at timestamptz NOT NULL,
    alias text CONSTRAINT developer_licenses_alias_key UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE developer_licenses;
-- +goose StatementEnd
