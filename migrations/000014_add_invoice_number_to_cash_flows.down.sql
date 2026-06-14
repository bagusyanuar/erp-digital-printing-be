DROP INDEX IF EXISTS idx_cash_flows_invoice_number;
ALTER TABLE cash_flows DROP COLUMN IF EXISTS invoice_number;
