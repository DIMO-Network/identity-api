-- +goose Up
-- +goose StatementBegin
ALTER TABLE rewards
    DROP CONSTRAINT rewards_synthetic_token_id,
    ADD CONSTRAINT rewards_synthetic_devices_id_fkey 
        FOREIGN KEY (synthetic_token_id)
        REFERENCES synthetic_devices (id)
        ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
    ALTER TABLE rewards 
        DROP CONSTRAINT rewards_synthetic_devices_id_fkey,
        ADD CONSTRAINT rewards_synthetic_token_id FOREIGN KEY (synthetic_token_id) REFERENCES synthetic_devices (id);
-- +goose StatementEnd
