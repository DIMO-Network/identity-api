-- +goose Up
-- +goose StatementBegin
ALTER TABLE manufacturers
    ADD COLUMN table_id int;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manufacturers
    DROP COLUMN table_id;
-- +goose StatementEnd
