-- +goose Up
-- +goose StatementBegin

-- Add templateId column to vehicle_sacds (nullable for backwards compatibility)
ALTER TABLE vehicle_sacds 
ADD COLUMN template_id bytea CHECK (length(template_id) = 32) NULL REFERENCES templates (id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicle_sacds DROP COLUMN template_id;
-- +goose StatementEnd