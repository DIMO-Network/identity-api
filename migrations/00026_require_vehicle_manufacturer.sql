-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles
    ALTER COLUMN manufacturer_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles
    ALTER COLUMN manufacturer_id SET NULL;
-- +goose StatementEnd
