-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts(
    kernel bytea CONSTRAINT accounts_kernel_address_check CHECK (length(kernel) = 20) PRIMARY KEY,
    signer bytea CONSTRAINT accounts_signer_check CHECK (length(signer) = 20) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    signer_added timestamptz NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE accounts;
-- +goose StatementEnd
