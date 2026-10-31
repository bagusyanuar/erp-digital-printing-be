ALTER TABLE orders 
DROP COLUMN IF EXISTS cancel_reason,
DROP COLUMN IF EXISTS cancelled_at;

