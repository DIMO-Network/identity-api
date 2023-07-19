-- +goose Up
-- +goose StatementBegin

CREATE TABLE aftermarket_devices(
    id int CONSTRAINT aftermarket_devices_pkey PRIMARY KEY,
    "address" bytea CONSTRAINT ad_address_check CHECK (length("address") = 20),
    "owner" bytea CONSTRAINT ad_owner_address_check CHECK (length("owner") = 20),
    "serial" text,
    imei text,
    minted_at timestamptz
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE aftermarket_devices;

-- +goose StatementEnd
