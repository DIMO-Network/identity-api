-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS unaccent;

ALTER TABLE manufacturers
    ADD COLUMN slug text CONSTRAINT manufacturers_slug_key UNIQUE;

UPDATE
    manufacturers
SET
    slug = lower(unaccent(replace(name, ' ', '-')));

ALTER TABLE manufacturers
    ALTER COLUMN slug SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manufacturers
    DROP COLUMN slug;
-- +goose StatementEnd
