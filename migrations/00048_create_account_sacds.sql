-- +goose Up
-- +goose StatementBegin
CREATE TABLE account_sacds (
    account bytea CONSTRAINT account_sacds_account_check CHECK (length(account) = 20),
    grantee bytea CONSTRAINT account_sacds_grantee_check CHECK (length(grantee) = 20),
    permissions bit varying(256) NOT NULL,
    source TEXT NOT NULL,
    template_id bytea CONSTRAINT account_sacds_template_id_check CHECK (length(template_id) = 32),
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT account_sacds_pkey PRIMARY KEY (account, grantee)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE account_sacds;
-- +goose StatementEnd
