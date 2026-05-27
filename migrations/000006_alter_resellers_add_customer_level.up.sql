-- Add customer_level_id to resellers
ALTER TABLE resellers ADD COLUMN customer_level_id UUID;

-- Backfill existing resellers to 'Reseller' customer level (UUID: d2c67ef8-82e4-4d8b-968b-5a1e2f5b6154)
UPDATE resellers SET customer_level_id = 'd2c67ef8-82e4-4d8b-968b-5a1e2f5b6154' WHERE customer_level_id IS NULL;

-- Add foreign key constraint
ALTER TABLE resellers ADD CONSTRAINT fk_resellers_customer_level FOREIGN KEY (customer_level_id) REFERENCES customer_levels(id) ON DELETE RESTRICT;
