-- +goose Up
-- +goose StatementBegin

SET search_path = identity_api, public;

CREATE TABLE aftermarket_devices(
    id numeric(78, 0),
    owner_address bytea
        CONSTRAINT minted_vehicles_owner_address_check CHECK (length(owner_address) = 20),
    beneficiary_address bytea
        CONSTRAINT device_beneficiary_address_check CHECK (length(beneficiary_address) = 20),
    vehicle_id numeric(78, 0),
    mint_time   timestamptz not null default current_timestamp,

    PRIMARY KEY (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SET search_path = identity_api, public;

drop table aftermarket_devices;

-- +goose StatementEnd
