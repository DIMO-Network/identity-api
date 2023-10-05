-- +goose Up
-- +goose StatementBegin
CREATE TABLE manufacturers(
    id int CONSTRAINT manufacturers_pkey PRIMARY KEY,
    name varchar NOT NULL,
    owner bytea NOT NULL CONSTRAINT manufacturers_owner_check CHECK (length(owner) = 20),
    minted_at timestamptz NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE manufacturers;
-- +goose StatementEnd
