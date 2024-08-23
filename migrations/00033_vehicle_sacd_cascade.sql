-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicle_sacds DROP CONSTRAINT vehicle_sacds_vehicle_id_fkey;

ALTER TABLE vehicle_sacds ADD CONSTRAINT vehicle_sacds_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES vehicles (id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicle_sacds DROP CONSTRAINT vehicle_sacds_vehicle_id_fkey;

ALTER TABLE vehicle_sacds ADD CONSTRAINT vehicle_sacds_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES vehicles (id);
-- +goose StatementEnd
