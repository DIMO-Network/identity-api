-- +goose Up
-- +goose StatementBegin
ALTER TABLE dcns ADD COLUMN vehicle_id int CONSTRAINT vehicle_dcn_vehicle_token_id REFERENCES vehicles (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE dcns DROP COLUMN vehicle_id;
-- +goose StatementEnd
