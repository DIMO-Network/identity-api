-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles
    ADD COLUMN image_uri text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles
    DROP COLUMN image_uri;
-- +goose StatementEnd
