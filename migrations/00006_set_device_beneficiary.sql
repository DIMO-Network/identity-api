-- +goose Up
-- +goose StatementBegin

ALTER TABLE aftermarket_devices
ADD COLUMN beneficiary bytea;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE aftermarket_devices
DROP COLUMN beneficiary;
-- +goose StatementEnd
