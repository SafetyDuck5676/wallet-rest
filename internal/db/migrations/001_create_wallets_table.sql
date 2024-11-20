CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL CHECK (balance >= 0),
    updated_at TIMESTAMP DEFAULT NOW()
);