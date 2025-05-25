CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_account BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    to_account BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    type VARCHAR(32) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    description TEXT,
    is_reversal BOOLEAN DEFAULT FALSE
)