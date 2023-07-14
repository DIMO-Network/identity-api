-- +goose Up
-- +goose StatementBegin
SET search_path = identity_api, public;

CREATE TABLE vehicles (
    id int CONSTRAINT vehicles_pkey PRIMARY KEY,
    owner_address bytea CONSTRAINT vehicles_owner_address_check CHECK (length(owner_address) = 20),
    make varchar(100),
    model varchar(100),
    year int,
    mint_time timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = identity_api, public;

DROP TABLE vehicles;
-- +goose StatementEnd