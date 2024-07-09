-- +goose Up
-- +goose StatementBegin
CREATE TABLE kernal_accounts(
    kernal bytea CONSTRAINT kernal_accounts_kernal_address_check CHECK (length(kernal) = 20) PRIMARY KEY,
    owner_address bytea CONSTRAINT kernal_accounts_owner_address_check CHECK (length(owner_address) = 20) NOT NULL -- TODO(ae): should this reference any other table?
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE kernal_accounts;
-- +goose StatementEnd
