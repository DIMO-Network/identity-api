-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles ADD COLUMN manufacturer_id int CONSTRAINT vehicles_manufacturer_id_fkey REFERENCES manufacturers(id);
ALTER TABLE aftermarket_devices ADD COLUMN manufacturer_id int CONSTRAINT aftermarket_devices_manufacturer_id_fkey REFERENCES manufacturers(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles DROP COLUMN manufacturer_id;
ALTER TABLE aftermarket_devices DROP COLUMN manufacturer_id;
-- +goose StatementEnd
