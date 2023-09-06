-- +goose Up
-- +goose StatementBegin
ALTER TABLE privileges
    DROP CONSTRAINT IF EXISTS privileges_vehicle_token_id_fkey,
    ADD CONSTRAINT privileges_vehicle_token_id_fkey FOREIGN KEY (token_id)
        REFERENCES vehicles (id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE privileges 
    DROP CONSTRAINT IF EXISTS privileges_vehicle_token_id_fkey,
    ADD CONSTRAINT privileges_vehicle_token_id_fkey REFERENCES vehicles (id);
-- +goose StatementEnd
