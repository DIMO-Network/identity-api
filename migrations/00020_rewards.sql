-- +goose Up
-- +goose StatementBegin
CREATE TABLE rewards (
    issuance_week int NOT NULL,
    connection_streak int,
    streak_earning int,
    vehicle_id int NOT NULL CONSTRAINT rewards_vehicle_token_id REFERENCES vehicles (id),
    aftermarket_token_id int CONSTRAINT rewards_aftermarket_token_id REFERENCES aftermarket_devices (id),
    aftermarket_earnings int,
    synthetic_token_id int CONSTRAINT rewards_synthetic_token_id REFERENCES synthetic_devices (id),
    synthetic_earnings int,
    received_by_address bytea CONSTRAINT received_by_address_check CHECK (length(received_by_address) = 20),
    earned_at timestamptz(0) NOT NULL,

    PRIMARY KEY (issuance_week, vehicle_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rewards;
-- +goose StatementEnd
