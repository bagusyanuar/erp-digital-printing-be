ALTER TABLE expense_categories ADD COLUMN sequence INT DEFAULT 0;

-- Auto-populate urutan expense_categories berdasarkan waktu buat
WITH numbered_categories AS (
    SELECT id, row_number() OVER (ORDER BY created_at ASC) as rn
    FROM expense_categories
)
UPDATE expense_categories
SET sequence = numbered_categories.rn
FROM numbered_categories
WHERE expense_categories.id = numbered_categories.id;
