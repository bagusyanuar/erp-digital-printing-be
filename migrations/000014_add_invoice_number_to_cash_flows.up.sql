ALTER TABLE cash_flows ADD COLUMN invoice_number VARCHAR(100);
CREATE INDEX idx_cash_flows_invoice_number ON cash_flows(invoice_number);
