CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL CHECK (balance >= 0),
    updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO wallets (id, balance, updated_at) 
VALUES ( 
    '19623ea0-ea8a-4b57-8303-1d81468a8f9d',   
    500,                                     
    '2024-11-20 12:00:00'                    
);