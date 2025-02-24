-- +goose Up
-- +goose StatementBegin
CREATE INDEX vehicles_owner_address_idx ON vehicles (owner_address);
CREATE INDEX privileges_user_address_idx ON privileges (user_address);
CREATE INDEX vehicle_sacds_grantee_idx ON vehicle_sacds (grantee);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX vehicle_sacds_grantee_idx;
DROP INDEX privileges_user_address_idx;
DROP INDEX vehicles_owner_address_idx;
-- +goose StatementEnd
