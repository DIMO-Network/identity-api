-- +goose Up
-- +goose StatementBegin
CREATE TABLE privileges (
    id         char(27) PRIMARY KEY,
    token_id int NOT NULL,
    privilege_id int NOT NULL,
    granted_to_address bytea CONSTRAINT granted_to_address_check CHECK (length(granted_to_address) = 20) NOT NULL,
    set_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,

    CONSTRAINT fkey_privileges_token_id FOREIGN KEY (token_id) REFERENCES vehicles(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE privileges;
-- +goose StatementEnd