-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles 
    ALTER COLUMN owner_address SET NOT NULL,
    ALTER COLUMN minted_at SET NOT NULL ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles 
    ALTER COLUMN owner_address DROP NOT NULL,
    ALTER COLUMN minted_at DROP NOT NULL;
-- +goose StatementEnd
