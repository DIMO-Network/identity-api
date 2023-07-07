-- +goose Up
-- +goose StatementBegin

CREATE TABLE vehicles(
    id numeric(78, 0),
    owner_address bytea
        CONSTRAINT minted_vehicles_owner_address_check CHECK (length(owner_address) = 20),
    make         varchar(100) not null,
    model        varchar(100) not null,
    year         int    not null,
    mint_time   timestamptz not null default current_timestamp,

    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table vehicles;

-- +goose StatementEnd
