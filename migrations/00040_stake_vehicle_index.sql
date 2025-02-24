-- +goose Up
-- +goose StatementBegin
ALTER TABLE stakes ADD CONSTRAINT stakes_vehicle_id_key UNIQUE (vehicle_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE stakes DROP CONSTRAINT stakes_vehicle_id_key;
-- +goose StatementEnd
