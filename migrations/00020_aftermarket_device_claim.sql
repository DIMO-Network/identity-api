-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ADD COLUMN claimed_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices DROP COLUMN claimed_at;
-- +goose StatementEnd
