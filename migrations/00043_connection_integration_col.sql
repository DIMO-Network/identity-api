-- +goose Up
-- +goose StatementBegin
-- This is the bytes32 resulting from bytes32(bytes(name)).
ALTER TABLE connections
    ADD COLUMN "id" bytea CHECK (length("id") = 32);

UPDATE
    connections
SET
    id = convert_to("name", 'UTF8') || decode(repeat('00', 32 - octet_length(name)), 'hex');

ALTER TABLE connections
    DROP CONSTRAINT connections_pkey;

ALTER TABLE connections
    ADD PRIMARY KEY (id);

ALTER TABLE connections
    DROP COLUMN "name";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE connections
    ADD COLUMN "name" text;

UPDATE
    connections
SET
    "name" = convert_from(rtrim(id, '\x00'::bytea), 'UTF8');

ALTER TABLE connections
    DROP CONSTRAINT connections_pkey;

ALTER TABLE connections
    ADD PRIMARY KEY (name);

ALTER TABLE connections
    DROP COLUMN id;
-- +goose StatementEnd
