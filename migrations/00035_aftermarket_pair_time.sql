-- +goose Up
-- +goose StatementBegin
-- TODO(elffjs): Make this nullable only when vehicle_token_id is also null.
ALTER TABLE aftermarket_devices ADD COLUMN paired_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices DROP COLUMN paired_at;
-- +goose StatementEnd
