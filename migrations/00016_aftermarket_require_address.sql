-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ALTER COLUMN address SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ALTER COLUMN address SET NULL;
-- +goose StatementEnd
