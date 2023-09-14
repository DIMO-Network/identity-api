-- +goose Up
-- +goose StatementBegin
ALTER TABLE dcns ADD COLUMN name varchar(200);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE dcns DROP COLUMN name;
-- +goose StatementEnd
