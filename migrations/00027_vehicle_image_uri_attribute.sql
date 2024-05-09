-- +goose Up
-- +goose StatementBegin
ALTER TABLE identity_api.vehicles ADD COLUMN image_uri text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE identity_api.vehicles DROP COLUMN image_uri;
-- +goose StatementEnd
