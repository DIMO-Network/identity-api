-- +goose Up
-- +goose StatementBegin
CREATE TABLE kernel_accounts(
    kernel bytea CONSTRAINT kernel_accounts_kernel_address_check CHECK (length(kernel) = 20) PRIMARY KEY,
    owner_address bytea CONSTRAINT kernel_accounts_owner_address_check CHECK (length(owner_address) = 20) NOT NULL -- TODO(ae): should this reference any other table?
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE kernel_accounts;
-- +goose StatementEnd
