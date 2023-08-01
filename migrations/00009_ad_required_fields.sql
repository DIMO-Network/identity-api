-- +goose Up
-- +goose StatementBegin

ALTER TABLE aftermarket_devices ALTER COLUMN "owner" SET NOT NULL;
ALTER TABLE aftermarket_devices ALTER COLUMN minted_at SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ALTER COLUMN "owner" DROP NOT NULL;
ALTER TABLE aftermarket_devices ALTER COLUMN minted_at DROP NOT NULL;
-- +goose StatementEnd
