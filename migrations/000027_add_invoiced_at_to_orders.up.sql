ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS invoiced_at TIMESTAMP WITH TIME ZONE;

-- Backfill existing invoiced orders using their first payment date or created_at
UPDATE orders 
SET invoiced_at = COALESCE(
    (SELECT MIN(created_at) FROM order_payments WHERE order_payments.order_id = orders.id AND order_payments.deleted_at IS NULL),
    orders.created_at
)
WHERE invoice_number IS NOT NULL AND invoiced_at IS NULL;

