-- +goose Up
-- +goose StatementBegin
ALTER TABLE identity_api.aftermarket_devices ADD COLUMN dev_eui text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE identity_api.aftermarket_devices DROP COLUMN dev_eui;
-- +goose StatementEnd
