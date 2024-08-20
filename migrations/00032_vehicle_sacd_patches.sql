-- +goose Up
-- +goose StatementBegin
ALTER TABLE sacds RENAME TO vehicle_sacds;

ALTER TABLE vehicle_sacds DROP CONSTRAINT sacds_pkey;

ALTER TABLE vehicle_sacds RENAME COLUMN token_id TO vehicle_id;

ALTER TABLE vehicle_sacds RENAME CONSTRAINT vehicle_privileges_vehicle_token_id TO vehicle_sacds_vehicle_id_fkey;
ALTER TABLE vehicle_sacds RENAME CONSTRAINT sacd_grantee_check TO vehicle_sacds_grantee_check;

ALTER TABLE vehicle_sacds ALTER COLUMN permissions TYPE bit varying(256);

ALTER TABLE vehicle_sacds ADD CONSTRAINT vehicle_sacds_pkey PRIMARY KEY (vehicle_id, grantee);

ALTER TABLE vehicle_sacds ALTER COLUMN source SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicle_sacds ALTER COLUMN source SET NOT NULL;

ALTER TABLE vehicle_sacds DROP CONSTRAINT vehicle_sacds_pkey;

ALTER TABLE vehicle_sacds ALTER COLUMN permissions TYPE bit(16);

ALTER TABLE vehicle_sacds RENAME CONSTRAINT vehicle_sacds_grantee_check TO sacd_grantee_check;
ALTER TABLE vehicle_sacds RENAME CONSTRAINT vehicle_sacds_vehicle_id_fkey TO vehicle_privileges_vehicle_token_id;

ALTER TABLE vehicle_sacds RENAME COLUMN vehicle_id TO token_id;

ALTER TABLE vehicle_sacds ADD CONSTRAINT sacds_pkey PRIMARY KEY (token_id, grantee, permissions);

ALTER TABLE vehicle_sacds RENAME TO sacd;
-- +goose StatementEnd
