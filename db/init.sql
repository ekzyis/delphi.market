CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ln_pubkey TEXT,
    nostr_pubkey TEXT,
    msats BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE sessions(
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_id VARCHAR(48)
);

CREATE TABLE lnauth(
    k1 VARCHAR(64) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    lnurl TEXT NOT NULL,
    session_id VARCHAR(48) NOT NULL DEFAULT encode(gen_random_uuid()::text::bytea, 'base64')
);

CREATE TABLE invoices(
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    msats BIGINT NOT NULL,
    msats_received BIGINT,
    preimage TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL UNIQUE,
    bolt11 TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    held_since TIMESTAMP WITH TIME ZONE,
    description TEXT
);

CREATE TABLE markets(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    description TEXT NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    settled_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    invoice_id INTEGER NOT NULL UNIQUE REFERENCES invoices(id)
);

CREATE TABLE shares(
    id SERIAL PRIMARY KEY,
    market_id INTEGER NOT NULL REFERENCES markets(id),
    description TEXT NOT NULL,
    win BOOLEAN
);

CREATE TYPE order_side AS ENUM ('BUY', 'SELL');

CREATE TABLE orders(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    share_id INTEGER NOT NULL REFERENCES shares(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    side ORDER_SIDE NOT NULL,
    quantity BIGINT NOT NULL,
    price BIGINT NOT NULL,
    invoice_id INTEGER REFERENCES invoices(id),
    order_id INTEGER REFERENCES orders(id)
);

ALTER TABLE orders ADD CONSTRAINT order_price CHECK(price > 0 AND price < 100);
ALTER TABLE orders ADD CONSTRAINT order_quantity CHECK(quantity > 0);

CREATE TABLE withdrawals(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    canceled_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    bolt11 TEXT NOT NULL UNIQUE,
    paid_at TIMESTAMP WITH TIME ZONE
);
