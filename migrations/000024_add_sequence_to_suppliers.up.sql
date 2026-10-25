ALTER TABLE suppliers ADD COLUMN sequence INT DEFAULT 0;

-- Auto-populate urutan supplier berdasarkan waktu buat
WITH numbered_suppliers AS (
    SELECT id, row_number() OVER (ORDER BY created_at ASC) as rn
    FROM suppliers
)
UPDATE suppliers
SET sequence = numbered_suppliers.rn
FROM numbered_suppliers
WHERE suppliers.id = numbered_suppliers.id;
