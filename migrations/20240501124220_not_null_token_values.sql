-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

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
SELECT 'down SQL query';
ALTER TABLE rewards
  ALTER COLUMN streak_earnings DROP DEFAULT,
  ALTER COLUMN streak_earnings DROP NOT NULL,
  ALTER COLUMN aftermarket_earnings DROP DEFAULT,
  ALTER COLUMN aftermarket_earnings DROP NOT NULL,
  ALTER COLUMN synthetic_earnings DROP DEFAULT,
  ALTER COLUMN synthetic_earnings DROP NOT NULL;
-- +goose StatementEnd
