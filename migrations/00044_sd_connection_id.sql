-- +goose Up
-- +goose StatementBegin
ALTER TABLE synthetic_devices ADD COLUMN connection_id bytea CHECK (length(connection_id) = 32) REFERENCES connections (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE synthetic_devices DROP COLUMN connection_id;
-- +goose StatementEnd
