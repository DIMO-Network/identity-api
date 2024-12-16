-- +goose Up
-- +goose StatementBegin
ALTER TABLE synthetic_devices ADD CONSTRAINT synthetic_devices_vehicle_id_key UNIQUE (vehicle_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE synthetic_devices DROP CONSTRAINT synthetic_devices_vehicle_id_key;
-- +goose StatementEnd
