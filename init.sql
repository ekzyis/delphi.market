CREATE TABLE lnauth(
    k1 VARCHAR(64) NOT NULL PRIMARY KEY,
    lnurl TEXT NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_id VARCHAR(48) NOT NULL DEFAULT encode(gen_random_uuid()::text::bytea, 'base64')
);
CREATE TABLE sessions(
    pubkey TEXT NOT NULL,
    session_id VARCHAR(48)
);
