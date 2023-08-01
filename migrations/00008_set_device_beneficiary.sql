-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices
    ADD COLUMN beneficiary bytea NOT NULL CONSTRAINT aftermarket_devices_beneficiary_check CHECK (length(beneficiary) = 20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices
    DROP COLUMN beneficiary;
-- +goose StatementEnd
