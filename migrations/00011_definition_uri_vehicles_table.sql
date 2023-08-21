-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicles ADD COLUMN definition_uri varchar(200);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles DROP COLUMN definition_uri;
-- +goose StatementEnd
