-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices
    ADD COLUMN vehicle_id int CONSTRAINT aftermarket_devices_vehicle_id_key UNIQUE CONSTRAINT aftermarket_devices_vehicle_id_fkey REFERENCES vehicles (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices
    DROP COLUMN vehicle_id;
-- +goose StatementEnd
