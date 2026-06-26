CREATE TABLE fund_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_date TIMESTAMP NOT NULL DEFAULT NOW(),
    from_account_id UUID NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    to_account_id UUID NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes TEXT,
    cashier_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_fund_transfers_deleted_at ON fund_transfers(deleted_at);
