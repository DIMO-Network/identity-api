-- +goose Up
-- +goose StatementBegin

ALTER TABLE aftermarket_devices
ADD COLUMN beneficiary bytea CONSTRAINT beneficiary_address_check CHECK (length(beneficiary) = 20) NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE aftermarket_devices
DROP COLUMN beneficiary;
-- +goose StatementEnd
