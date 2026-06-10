-- +goose Up
-- +goose StatementBegin
CREATE TABLE merkle_pools (
    pool_id int CONSTRAINT merkle_pools_pkey PRIMARY KEY,
    token bytea NOT NULL CONSTRAINT merkle_pools_token_check CHECK (length(token) = 20),
    admin bytea NOT NULL CONSTRAINT merkle_pools_admin_check CHECK (length(admin) = 20),
    weekly_limit numeric(38, 0),
    balance numeric(38, 0) NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL
);

CREATE TABLE merkle_roots (
    pool_id int NOT NULL CONSTRAINT merkle_roots_pool_id_fkey REFERENCES merkle_pools (pool_id),
    epoch int NOT NULL,
    root bytea NOT NULL CONSTRAINT merkle_roots_root_check CHECK (length(root) = 32),
    allocation numeric(38, 0) NOT NULL,
    total_claimed numeric(38, 0) NOT NULL DEFAULT 0,
    claim_count int NOT NULL DEFAULT 0,
    recipient_count int NOT NULL DEFAULT 0,
    proofs_uri text NOT NULL,
    set_at timestamptz NOT NULL,
    CONSTRAINT merkle_roots_pkey PRIMARY KEY (pool_id, epoch)
);

CREATE TABLE merkle_claims (
    pool_id int NOT NULL,
    epoch int NOT NULL,
    account bytea NOT NULL CONSTRAINT merkle_claims_account_check CHECK (length(account) = 20),
    amount numeric(38, 0) NOT NULL,
    proof jsonb NOT NULL,
    claimed_at timestamptz,
    claim_tx bytea CONSTRAINT merkle_claims_claim_tx_check CHECK (claim_tx IS NULL OR length(claim_tx) = 32),
    CONSTRAINT merkle_claims_pkey PRIMARY KEY (pool_id, epoch, account),
    CONSTRAINT merkle_claims_pool_id_epoch_fkey FOREIGN KEY (pool_id, epoch) REFERENCES merkle_roots (pool_id, epoch)
);

CREATE INDEX merkle_claims_account_idx ON merkle_claims (account);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE merkle_claims;

DROP TABLE merkle_roots;

DROP TABLE merkle_pools;
-- +goose StatementEnd
