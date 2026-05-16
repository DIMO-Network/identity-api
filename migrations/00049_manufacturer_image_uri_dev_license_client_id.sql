-- +goose Up
-- +goose StatementBegin
ALTER TABLE manufacturers
    ADD COLUMN image_uri              TEXT,
    ADD COLUMN dev_license_client_id  BYTEA
        CONSTRAINT manufacturers_dev_license_client_id_check
        CHECK (dev_license_client_id IS NULL OR length(dev_license_client_id) = 20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manufacturers
    DROP COLUMN dev_license_client_id,
    DROP COLUMN image_uri;
-- +goose StatementEnd
