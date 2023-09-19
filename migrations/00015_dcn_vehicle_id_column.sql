-- +goose Up
-- +goose StatementBegin
ALTER TABLE dcns ADD COLUMN vehicle_id int CONSTRAINT dcns_vehicle_id_fkey REFERENCES vehicles (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE dcns DROP COLUMN vehicle_id;
-- +goose StatementEnd
