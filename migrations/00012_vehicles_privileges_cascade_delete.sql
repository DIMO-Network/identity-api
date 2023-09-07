-- +goose Up
-- +goose StatementBegin
ALTER TABLE privileges
    DROP CONSTRAINT vehicle_privileges_vehicle_token_id,
    ADD CONSTRAINT privileges_vehicle_token_id_fkey FOREIGN KEY (token_id)
        REFERENCES vehicles (id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE privileges 
    DROP CONSTRAINT privileges_vehicle_token_id_fkey,
    ADD CONSTRAINT vehicle_privileges_vehicle_token_id REFERENCES vehicles (id);
-- +goose StatementEnd
