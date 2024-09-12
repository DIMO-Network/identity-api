-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ADD COLUMN hardware_revision TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices DROP COLUMN hardware_revision;
-- +goose StatementEnd
