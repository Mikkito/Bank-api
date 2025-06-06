CREATE TABLE IF NOT EXISTS loans (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    principal NUMERIC(12, 2) NOT NULL,
    interest_rate NUMERIC(5, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_repaid BOOLEAN NOT NULL DEFAULT FALSE,
    start_date TIMESTAMP NOT NULL DEFAULT NOW(),
    next_payment_due TIMESTAMP
);