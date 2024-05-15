-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles
    ADD COLUMN device_definition_id text,
    DROP COLUMN definition_uri;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles
    ADD COLUMN definition_uri varchar(200),
    DROP COLUMN device_definition;
-- +goose StatementEnd
