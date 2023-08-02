-- +goose Up
-- +goose StatementBegin

CREATE TABLE aftermarket_devices(
    id int CONSTRAINT aftermarket_devices_pkey PRIMARY KEY,
    "address" bytea CONSTRAINT aftermarket_devices_address_check CHECK (length("address") = 20),
    "owner" bytea CONSTRAINT aftermarket_devices_owner_check CHECK (length("owner") = 20),
    "serial" text,
    imei text,
    minted_at timestamptz(0)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE aftermarket_devices;

-- +goose StatementEnd
