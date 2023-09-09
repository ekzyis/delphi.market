CREATE TABLE lnauth(
    k1 VARCHAR(64) NOT NULL PRIMARY KEY,
    lnurl TEXT NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_id VARCHAR(48) NOT NULL DEFAULT encode(gen_random_uuid()::text::bytea, 'base64')
);
CREATE TABLE users(
    pubkey TEXT PRIMARY KEY,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE sessions(
    pubkey TEXT NOT NULL REFERENCES users(pubkey),
    session_id VARCHAR(48)
);

CREATE TABLE markets(
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    funding BIGINT NOT NULL,
    active BOOLEAN DEFAULT true
);
CREATE EXTENSION "uuid-ossp";
CREATE TABLE contracts(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    market_id INTEGER REFERENCES markets(id),
    description TEXT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL
);
CREATE TYPE order_side AS ENUM ('BUY', 'SELL');
CREATE TABLE trades(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id),
    pubkey TEXT NOT NULL REFERENCES users(pubkey),
    side ORDER_SIDE NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    msats BIGINT NOT NULL
);
