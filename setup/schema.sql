CREATE TABLE IF NOT EXISTS blocks (
    id VARCHAR PRIMARY KEY,
    block_number VARCHAR,
    state VARCHAR,
    chain_id VARCHAR
);

CREATE TABLE IF NOT EXISTS blocks_pointers (
    id VARCHAR PRIMARY KEY,
    chain_id VARCHAR,
    block_number VARCHAR
);

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR PRIMARY KEY,
    chain_id VARCHAR NOT NULL,
    state VARCHAR NOT NULL,
    block_number VARCHAR NOT NULL,
    creation_date BIGINT NOT NULL,
    last_updated BIGINT NOT NULL
);
