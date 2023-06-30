-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = identity_api, public;

CREATE TABLE minted_vehicles(
    id numeric(78, 0)
        CONSTRAINT minted_vehicles_id_key UNIQUE,
    owner_address bytea
        CONSTRAINT minted_vehicles_owner_address_check CHECK (length(owner_address) = 20),
    make         varchar(100) not null,
    model        varchar(100) not null,
    year         smallint    not null,
    mint_time   timestamptz not null default current_timestamp,

    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

drop table minted_vehicles;

-- +goose StatementEnd
