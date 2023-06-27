-- +goose Up
-- +goose StatementBegin
REVOKE CREATE ON schema public FROM public;
CREATE SCHEMA IF NOT EXISTS identity_api;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA identity_api CASCADE;
GRANT CREATE, USAGE ON schema public TO public;
-- +goose StatementEnd
