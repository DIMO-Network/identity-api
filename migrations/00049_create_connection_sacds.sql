-- +goose Up
-- +goose StatementBegin
CREATE TABLE connection_sacds (
    connection_id bytea CONSTRAINT connection_sacds_connection_id_check CHECK (length(connection_id) = 32)
        REFERENCES connections (id) ON DELETE CASCADE,
    grantee bytea CONSTRAINT connection_sacds_grantee_check CHECK (length(grantee) = 20),
    permissions bit varying(256) NOT NULL,
    source TEXT NOT NULL,
    template_id bytea CONSTRAINT connection_sacds_template_id_check CHECK (length(template_id) = 32),
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT connection_sacds_pkey PRIMARY KEY (connection_id, grantee)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE connection_sacds;
-- +goose StatementEnd
