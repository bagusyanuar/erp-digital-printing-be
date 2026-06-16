-- 1. Hapus FK lama & kolom lama di expenses
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS fk_expenses_expense_category;
ALTER TABLE expenses DROP COLUMN IF EXISTS expense_category_id;
ALTER TABLE expenses DROP COLUMN IF EXISTS payment_method;

-- 2. Tambah kolom baru di expenses
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS expense_number VARCHAR(100) UNIQUE;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100);
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS vendor_name VARCHAR(255) NOT NULL DEFAULT 'Umum';
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'PAID';

-- 3. Buat tabel expense_items
CREATE TABLE IF NOT EXISTS expense_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    description VARCHAR(255),
    qty INT NOT NULL DEFAULT 1,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_expense_items_deleted_at ON expense_items(deleted_at);

-- 4. Buat tabel expense_payments
CREATE TABLE IF NOT EXISTS expense_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    payment_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    payment_method VARCHAR(50) NOT NULL,
    cashier_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_expense_payments_deleted_at ON expense_payments(deleted_at);

-- 5. Seed Kategori Khusus Diskon
INSERT INTO expense_categories (id, name, "group", created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000000', 'Potongan Pembelian', 'OPERATIONAL', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;
