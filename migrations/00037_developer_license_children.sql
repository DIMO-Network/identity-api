-- +goose Up
-- +goose StatementBegin
CREATE TABLE redirect_uris(
    developer_license_id int NOT NULL CONSTRAINT redirect_uris_developer_license_id_fkey REFERENCES developer_licenses (id) ON DELETE CASCADE,
    uri text NOT NULL,
    enabled_at timestamptz NOT NULL,
    CONSTRAINT redirect_uris_pkey PRIMARY KEY (developer_license_id, uri)
);

CREATE TABLE signers(
    developer_license_id int NOT NULL CONSTRAINT signers_developer_license_id_fkey REFERENCES developer_licenses (id) ON DELETE CASCADE,
    signer bytea NOT NULL CONSTRAINT signers_signer CHECK (length(signer) = 20),
    enabled_at timestamptz NOT NULL,
    CONSTRAINT signers_pkey PRIMARY KEY (developer_license_id, signer)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE signers;
DROP TABLE redirect_uris;
-- +goose StatementEnd
