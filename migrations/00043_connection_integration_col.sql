-- +goose Up
-- +goose StatementBegin
ALTER TABLE connections
    ADD CONSTRAINT CHECK (octet_length(name) <= 32)
    ADD COLUMN integration_id int;

ALTER TABLE synthetic_devices
    ADD COLUMN connection_name text CHECK (octet_length(connection_name) <= 32) REFERENCES connections (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE synthetic_devices
    DROP COLUMN connection_address;

ALTER TABLE connections
    DROP COLUMN integration_id;
-- +goose StatementEnd
