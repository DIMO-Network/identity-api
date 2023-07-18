-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles RENAME COLUMN mint_time TO minted_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles RENAME COLUMN minted_at TO mint_time;
-- +goose StatementEnd
