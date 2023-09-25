-- +goose Up
-- +goose StatementBegin
ALTER TABLE dcns ADD COLUMN minted_at timestamptz NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE dcns DROP COLUMN minted_at;
-- +goose StatementEnd
