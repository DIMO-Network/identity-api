-- +goose Up
-- +goose StatementBegin

ALTER TABLE aftermarket_devices
ADD COLUMN vehicle_id int;

ALTER TABLE aftermarket_devices
ADD CONSTRAINT ad_vehicle_link FOREIGN KEY (vehicle_id) REFERENCES vehicles (id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE aftermarket_devices
DROP CONSTRAINT ad_vehicle_link;

ALTER TABLE aftermarket_devices
DROP COLUMN vehicle_id;
-- +goose StatementEnd
