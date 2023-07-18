-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles ALTER COLUMN owner_address SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles ALTER COLUMN owner_address DROP NOT NULL;
-- +goose StatementEnd
