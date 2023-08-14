-- +goose Up
-- +goose StatementBegin
CREATE TABLE synthetic_devices (
    id int PRIMARY KEY,
    integration_id int NOT NULL,
    vehicle_id int NOT NULL CONSTRAINT synthetic_devices_vehicle_token_id REFERENCES vehicles (id),
    device_address bytea CONSTRAINT device_address_check CHECK (length(device_address) = 20) NOT NULL,
    minted_at timestamptz(0) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE synthetic_devices;
-- +goose StatementEnd
