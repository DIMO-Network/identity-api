-- +goose Up
-- +goose StatementBegin
UPDATE
    rewards
SET
    streak_earnings = COALESCE(streak_earnings, 0),
    aftermarket_earnings = COALESCE(aftermarket_earnings, 0),
    synthetic_earnings = COALESCE(synthetic_earnings, 0);

ALTER TABLE rewards
    ALTER COLUMN streak_earnings SET DEFAULT 0,
    ALTER COLUMN streak_earnings SET NOT NULL,
    ALTER COLUMN aftermarket_earnings SET DEFAULT 0,
    ALTER COLUMN aftermarket_earnings SET NOT NULL,
    ALTER COLUMN synthetic_earnings SET DEFAULT 0,
    ALTER COLUMN synthetic_earnings SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rewards
    ALTER COLUMN streak_earnings DROP DEFAULT,
    ALTER COLUMN streak_earnings DROP NOT NULL,
    ALTER COLUMN aftermarket_earnings DROP DEFAULT,
    ALTER COLUMN aftermarket_earnings DROP NOT NULL,
    ALTER COLUMN synthetic_earnings DROP DEFAULT,
    ALTER COLUMN synthetic_earnings DROP NOT NULL;
-- +goose StatementEnd
