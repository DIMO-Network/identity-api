-- +goose Up
-- +goose StatementBegin
CREATE TABLE vehicles (
    id int CONSTRAINT vehicles_pkey PRIMARY KEY,
    owner_address bytea CONSTRAINT vehicles_owner_address_check CHECK (length(owner_address) = 20),
    make varchar(100),
    model varchar(100),
    year int,
    mint_time timestamptz(0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE vehicles;
-- +goose StatementEnd
