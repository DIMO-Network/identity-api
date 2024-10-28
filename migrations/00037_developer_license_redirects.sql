-- +goose Up
-- +goose StatementBegin
CREATE TABLE redirect_uris(
    developer_license_id int NOT NULL CONSTRAINT redirect_uris_developer_license_id_fkey REFERENCES developer_licenses (id) ON DELETE CASCADE,
    uri text NOT NULL,
    enabled_at timestamptz NOT NULL,
    CONSTRAINT redirect_uris_pkey PRIMARY KEY (developer_license_id, uri)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE redirect_uris;
-- +goose StatementEnd
