-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles ALTER COLUMN minted_at SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles ALTER COLUMN minted_at DROP NOT NULL;
-- +goose StatementEnd
