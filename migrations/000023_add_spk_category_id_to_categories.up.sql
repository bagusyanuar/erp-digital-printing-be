ALTER TABLE categories ADD COLUMN IF NOT EXISTS spk_category_id UUID REFERENCES categories(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_categories_spk_category_id ON categories(spk_category_id);
