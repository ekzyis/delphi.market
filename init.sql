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
    active BOOLEAN DEFAULT true
);
CREATE EXTENSION "uuid-ossp";
CREATE TABLE shares(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    market_id INTEGER REFERENCES markets(id),
    description TEXT NOT NULL
);
CREATE TYPE order_side AS ENUM ('BUY', 'SELL');
CREATE TABLE orders(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    share_id UUID NOT NULL REFERENCES shares(id),
    pubkey TEXT NOT NULL REFERENCES users(pubkey),
    side ORDER_SIDE NOT NULL,
    quantity BIGINT NOT NULL,
    price BIGINT NOT NULL
);
CREATE TABLE trades(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id_1 UUID NOT NULL REFERENCES orders(id),
    order_id_2 UUID NOT NULL REFERENCES orders(id)
);
