-- +goose Up
-- +goose StatementBegin

CREATE TABLE vehicles(
    id int,
    owner_address bytea null
        CONSTRAINT minted_vehicles_owner_address_check CHECK (length(owner_address) = 20),
    make         varchar(100) null,
    model        varchar(100) null,
    year         int    null,
    mint_time   timestamptz null,

    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table vehicles;

-- +goose StatementEnd
