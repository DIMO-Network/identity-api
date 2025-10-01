-- +goose Up
-- +goose StatementBegin

-- Create templates table for the new TemplateCreated event
CREATE TABLE templates (
    id bytea PRIMARY KEY CHECK (length(id) = 32), -- This is the keccak256 of cid.
    creator bytea NOT NULL CHECK (length(creator) = 20),
    asset bytea NOT NULL CHECK (length(asset) = 20),
    permissions bit varying(256) NOT NULL,
    cid TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE templates;
-- +goose StatementEnd
