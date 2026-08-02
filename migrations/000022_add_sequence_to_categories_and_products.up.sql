ALTER TABLE categories ADD COLUMN sequence INT DEFAULT 0;
ALTER TABLE products ADD COLUMN sequence INT DEFAULT 0;

-- Auto-populate urutan category berdasarkan waktu buat
WITH numbered_categories AS (
    SELECT id, row_number() OVER (ORDER BY created_at ASC) as rn
    FROM categories
)
UPDATE categories
SET sequence = numbered_categories.rn
FROM numbered_categories
WHERE categories.id = numbered_categories.id;

-- Auto-populate urutan product berdasarkan waktu buat
WITH numbered_products AS (
    SELECT id, row_number() OVER (ORDER BY created_at ASC) as rn
    FROM products
)
UPDATE products
SET sequence = numbered_products.rn
FROM numbered_products
WHERE products.id = numbered_products.id;
