-- +goose Up
-- +goose StatementBegin
CREATE TABLE stakes (
    id int CONSTRAINT stakes_pkey PRIMARY KEY,
    owner bytea NOT NULL CONSTRAINT stakes_owner_check CHECK (length(owner) = 20),
    level int NOT NULL,
    points int NOT NULL,
    amount numeric(30, 0) NOT NULL,
    vehicle_id int CONSTRAINT stakes_vehicle_id_fkey REFERENCES vehicles (id) ON DELETE SET NULL,
    staked_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    withdrawn_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stakes;
-- +goose StatementEnd
