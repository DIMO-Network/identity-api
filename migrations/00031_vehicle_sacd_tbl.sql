-- +goose Up
-- +goose StatementBegin
CREATE TABLE sacds (
    token_id int NOT NULL CONSTRAINT vehicle_privileges_vehicle_token_id REFERENCES vehicles (id),
    grantee bytea NOT NULL CONSTRAINT sacd_grantee_check CHECK (length(grantee) = 20),
    permissions bit(16) NOT NULL,
    source TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (token_id, grantee, permissions)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sacds;
-- +goose StatementEnd
