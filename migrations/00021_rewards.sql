-- +goose Up
-- +goose StatementBegin
CREATE TABLE rewards (
    issuance_week int NOT NULL,
    vehicle_id int CONSTRAINT rewards_vehicle_token_id REFERENCES vehicles (id),
    connection_streak int,
    streak_earnings numeric(30,0),
    aftermarket_token_id int CONSTRAINT rewards_aftermarket_token_id REFERENCES aftermarket_devices (id),
    aftermarket_earnings numeric(30,0),
    synthetic_token_id int CONSTRAINT rewards_synthetic_token_id REFERENCES synthetic_devices (id),
    synthetic_earnings numeric(30,0),
    received_by_address bytea CONSTRAINT received_by_address_check CHECK (length(received_by_address) = 20),
    earned_at timestamptz(0) NOT NULL,

    PRIMARY KEY (issuance_week, vehicle_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rewards;
-- +goose StatementEnd
