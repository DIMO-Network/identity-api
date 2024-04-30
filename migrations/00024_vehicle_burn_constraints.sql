-- +goose Up
-- +goose StatementBegin
ALTER TABLE rewards
    DROP CONSTRAINT rewards_vehicle_token_id,
    ADD CONSTRAINT rewards_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES vehicles (id) ON DELETE CASCADE;

ALTER TABLE dcns
    DROP CONSTRAINT vehicle_dcn_vehicle_token_id,
    ADD CONSTRAINT dcns_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES vehicles (id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rewards
    DROP CONSTRAINT rewards_vehicle_id_fkey,
    ADD CONSTRAINT rewards_vehicle_token_id FOREIGN KEY (vehicle_id) REFERENCES vehicles (id);

ALTER TABLE dcns
    DROP CONSTRAINT dcns_vehicle_id_fkey,
    ADD CONSTRAINT vehicle_dcn_vehicle_token_id FOREIGN KEY (vehicle_id) REFERENCES vehicles (id);
-- +goose StatementEnd
