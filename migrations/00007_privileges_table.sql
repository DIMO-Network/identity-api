-- +goose Up
-- +goose StatementBegin
CREATE TABLE privileges (
    token_id int NOT NULL CONSTRAINT vehicle_privileges_vehicle_token_id REFERENCES vehicles (id),
    privilege_id int NOT NULL,
    user_address bytea CONSTRAINT vehicle_privileges_user_address_check CHECK (length(user_address) = 20) NOT NULL,
    set_at timestamptz(0) NOT NULL,
    expires_at timestamptz(0) NOT NULL,

    PRIMARY KEY (token_id, privilege_id, user_address)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE privileges;
-- +goose StatementEnd
