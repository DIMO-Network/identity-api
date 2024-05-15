-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles
    ADD COLUMN device_definition_id text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles
    DROP COLUMN device_definition;
-- +goose StatementEnd
