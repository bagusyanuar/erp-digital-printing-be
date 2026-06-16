DROP TABLE IF EXISTS expense_payments;
DROP TABLE IF EXISTS expense_items;

ALTER TABLE expenses DROP COLUMN IF EXISTS status;
ALTER TABLE expenses DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE expenses DROP COLUMN IF EXISTS supplier_id;
ALTER TABLE expenses DROP COLUMN IF EXISTS invoice_number;
ALTER TABLE expenses DROP COLUMN IF EXISTS expense_number;

ALTER TABLE expenses ADD COLUMN expense_category_id UUID REFERENCES expense_categories(id) ON DELETE RESTRICT;
ALTER TABLE expenses ADD COLUMN payment_method VARCHAR(50);

DELETE FROM expense_categories WHERE id = '00000000-0000-0000-0000-000000000000';
