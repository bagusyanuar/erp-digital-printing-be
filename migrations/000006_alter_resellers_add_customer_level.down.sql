ALTER TABLE resellers DROP CONSTRAINT IF EXISTS fk_resellers_customer_level;
ALTER TABLE resellers DROP COLUMN IF EXISTS customer_level_id;
