CREATE TABLE capital_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_date TIMESTAMP NOT NULL DEFAULT NOW(),
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(50) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_capital_transactions_deleted_at ON capital_transactions(deleted_at);
