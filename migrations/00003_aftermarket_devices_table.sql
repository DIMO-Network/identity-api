-- +goose Up
-- +goose StatementBegin

SET search_path = identity_api, public;

CREATE TABLE aftermarket_devices(
    id int CONSTRAINT aftermarket_devices_pkey PRIMARY KEY,
    "address" bytea not null,
    "owner" bytea not null,
    "serial" text,
    imei text,
    minted_at timestamptz not null,
    vehicle_id int,

    CONSTRAINT linked_ad_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = identity_api, public;

drop table aftermarket_devices;

-- +goose StatementEnd
