CREATE INDEX IF NOT EXISTS idx_order_payments_order_id ON order_payments (order_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_order_payments_refund ON order_payments (order_id) WHERE payment_type = 'REFUND' AND deleted_at IS NULL;

