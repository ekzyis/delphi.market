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
CREATE TABLE invoices(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pubkey TEXT NOT NULL REFERENCES users(pubkey),
    msats BIGINT NOT NULL,
    msats_received BIGINT,
    preimage TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL UNIQUE,
    bolt11 TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    held_since TIMESTAMP WITH TIME ZONE
);
CREATE TYPE order_side AS ENUM ('BUY', 'SELL');
CREATE TABLE orders(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    share_id UUID NOT NULL REFERENCES shares(id),
    pubkey TEXT NOT NULL REFERENCES users(pubkey),
    side ORDER_SIDE NOT NULL,
    quantity BIGINT NOT NULL,
    price BIGINT NOT NULL,
    invoice_id UUID NOT NULL REFERENCES invoices(id)
);
ALTER TABLE orders ADD CONSTRAINT order_price CHECK(price > 0 AND price < 100);
CREATE TABLE matches(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    oid1 UUID NOT NULL REFERENCES orders(id),
    oid2 UUID NOT NULL REFERENCES orders(id)
);