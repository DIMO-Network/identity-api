-- +goose Up
-- +goose StatementBegin
ALTER TABLE synthetic_devices ADD COLUMN connection_id bytea CHECK (length(connection_id) = 32) REFERENCES connections (id);

ALTER TABLE connections ADD COLUMN integration_node int UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE connections DROP COLUMN integration_node;

ALTER TABLE synthetic_devices DROP COLUMN connection_id;
-- +goose StatementEnd
