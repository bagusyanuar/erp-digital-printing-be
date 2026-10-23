DROP INDEX IF EXISTS idx_categories_spk_category_id;
ALTER TABLE categories DROP COLUMN IF EXISTS spk_category_id;
