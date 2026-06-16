# Planning: Pencatatan Pengeluaran dengan Skema Termin / Hutang

Dokumen ini menjelaskan rancangan database, alur kerja, dan desain API untuk pencatatan pengeluaran (expenses) yang mendukung:
- Pembayaran termin (hutang/cicilan).
- Integrasi supplier (terdaftar & tidak terdaftar).
- Multi-item per nota (header-detail pattern).
- Diskon di level item & nota.
- Integrasi dengan cash flow.

---

## 1. Desain Database

### A. Tabel `suppliers` (Master Data Supplier)
Menyimpan data supplier yang terdaftar.
* `id` (UUID, PK)
* `name` (VARCHAR, Not Null)
* `contact_name` (VARCHAR, Nullable)
* `phone` (VARCHAR, Nullable)
* `email` (VARCHAR, Nullable)
* `address` (TEXT, Nullable)
* `created_at`, `updated_at`, `deleted_at`

---

### B. Tabel `expenses` (Header Nota)
Mencatat informasi utama dari nota pengeluaran, status pembayaran, dan relasi ke supplier.
* `id` (UUID, PK)
* `expense_number` (VARCHAR, Unique, Not Null) — Nomor referensi internal (generate otomatis, e.g., `EXP/20260616/0001`).
* `invoice_number` (VARCHAR, Nullable) — Nomor nota/faktur dari supplier (opsional, tidak semua pengeluaran punya nota fisik).
* `supplier_id` (UUID, FK → suppliers, Nullable) — ID Supplier jika terdaftar. NULL jika non-supplier.
* `vendor_name` (VARCHAR, Not Null) — Nama vendor/supplier (snapshot nama supplier terdaftar, atau teks bebas jika non-supplier).
* `amount` (DECIMAL(15,2)) — Total tagihan bersih (`SUM(expense_items.amount)`). Ini yang jadi acuan pembayaran.
* `status` (VARCHAR) — `'PAID'`, `'PARTIAL'`, `'UNPAID'`.
* `due_date` (TIMESTAMP, Nullable) — Jatuh tempo, wajib diisi jika `PARTIAL` or `UNPAID`.
* `expense_date` (TIMESTAMP) — Tanggal transaksi.
* `description` (TEXT, Nullable) — Catatan tambahan.
* `cashier_id` (UUID, Not Null) — ID kasir yang menginput.
* `created_at`, `updated_at`, `deleted_at`

---

### C. Tabel `expense_items` (Detail Item Per Nota)
Mencatat setiap baris item pengeluaran dalam 1 nota. Diskon dicatat sebagai item dengan nominal negatif menggunakan kategori khusus "Potongan Pembelian".
* `id` (UUID, PK)
* `expense_id` (UUID, FK → expenses, Not Null)
* `expense_category_id` (UUID, FK → expense_categories, Not Null) — Kategori pengeluaran per item.
* `description` (VARCHAR, Nullable) — Keterangan item (e.g., "Kertas Art Paper 100gsm" atau "Diskon Supplier").
* `qty` (INTEGER, Default 1) — Jumlah.
* `price` (DECIMAL(15,2)) — Harga satuan (bisa bernilai negatif untuk item diskon/potongan).
* `amount` (DECIMAL(15,2)) — Total per item (`qty × price`).
* `created_at`, `updated_at`, `deleted_at`

---

### D. Tabel `expense_payments` (Pembayaran / Cicilan)
Mencatat setiap transaksi pembayaran terhadap suatu nota pengeluaran.
* `id` (UUID, PK)
* `expense_id` (UUID, FK → expenses, Not Null)
* `amount` (DECIMAL(15,2)) — Jumlah yang dibayarkan.
* `payment_date` (TIMESTAMP) — Tanggal pembayaran.
* `payment_method` (VARCHAR) — `'CASH'`, `'TRANSFER'`, `'GIRO'`, dll.
* `cashier_id` (UUID, Not Null) — ID kasir yang memproses.
* `created_at`, `updated_at`, `deleted_at`

---

## 2. Kalkulasi Diskon (Metode Item Negatif)

Diskon tidak disimpan di kolom khusus, melainkan diinput sebagai baris pengeluaran bernilai negatif menggunakan kategori khusus "Potongan Pembelian":

```
expense_items:
  Item 1 (Art Paper)         : qty(10) × price(300.000)  =  3.000.000
  Item 2 (Toner)             : qty(5)  × price(400.000)  =  2.000.000
  Item 3 (Potongan Pembelian): qty(1)  × price(-500.000) = -  500.000

expenses:
  amount = SUM(items.amount) = 3.000.000 + 2.000.000 + (-500.000) = 4.500.000  ← total tagihan
```

---

## 3. Alur Pengisian (Workflow)

### A. Input Vendor/Supplier
1. **Supplier Terdaftar**: Pilih dropdown → `supplier_id` = UUID, `vendor_name` = nama snapshot.
2. **Non-Supplier**: Ketik manual → `supplier_id` = NULL, `vendor_name` = teks input.

### B. Input Pembayaran
1. **PAID (Lunas)**: Otomatis buat record `expense_payments` senilai pembayaran yang diinput di array `payments` (total `payments.amount` == `expenses.amount`).
2. **UNPAID (Hutang Penuh)**: Array `payments` kosong.
3. **PARTIAL (Termin/DP)**: Buat record `expense_payments` senilai DP yang diinput di array `payments`.

### C. Pembayaran Cicilan Berikutnya
1. Jalankan DB Transaction (`tx`).
2. Hitung total bayar: `SUM(expense_payments.amount) + SUM(payments_baru.amount)`.
3. Validasi: total tidak boleh melebihi `expenses.amount`.
4. Simpan record baru ke `expense_payments`.
5. Update `expenses.status`:
   - Total bayar == `expenses.amount` → `'PAID'`.
   - Total bayar < `expenses.amount` → `'PARTIAL'`.
6. Commit.

---

## 4. Integrasi Cash Flow

### Prinsip
Cash flow **hanya mencatat uang yang benar-benar keluar**. Diskon tidak muncul di cash_flow.

### Mapping
| `expense_payments` | → | `cash_flows` |
|---|---|---|
| `id` | → | `reference_id` |
| — | → | `reference_type = 'EXPENSE_PAYMENT'` |
| `amount` | → | `amount` |
| — | → | `type = 'CREDIT'` (uang keluar) |
| `payment_method` | → | `payment_method` |
| `payment_date` | → | `transaction_date` |
| `cashier_id` | → | `cashier_id` |

### Contoh
```
Nota: Rp 5.000.000 (gross) - Diskon Rp 500.000 = Rp 4.500.000

Payment 1 (DP Split - Cash & Transfer):
  expense_payments 1 → amount: 1.000.000, method: 'CASH'
  cash_flows 1       → CREDIT 1.000.000, method: 'CASH', ref_type: 'EXPENSE_PAYMENT'
  
  expense_payments 2 → amount: 1.000.000, method: 'TRANSFER'
  cash_flows 2       → CREDIT 1.000.000, method: 'TRANSFER', ref_type: 'EXPENSE_PAYMENT'

Total uang keluar = 2.000.000 ✅
```

---

## 5. Desain API

### A. Create Expense
`POST /api/v1/expenses`
```json
{
  "invoice_number": "INV/2026/XYZ",
  "supplier_id": "uuid-atau-null",
  "vendor_name": "PT. Kertas Jaya",
  "expense_date": "2026-06-16T12:00:00Z",
  "description": "Pembelian bahan cetak termin 1 bulan",
  "items": [
    {
      "expense_category_id": "uuid-kategori-art-paper",
      "description": "Kertas Art Paper 100gsm",
      "qty": 10,
      "price": 300000
    },
    {
      "expense_category_id": "uuid-kategori-toner",
      "description": "Tinta DTF",
      "qty": 5,
      "price": 400000
    },
    {
      "expense_category_id": "uuid-kategori-potongan-pembelian",
      "description": "Diskon Khusus Supplier",
      "qty": 1,
      "price": -500000
    }
  ],
  "payments": [
    {
      "amount": 1000000,
      "payment_method": "CASH"
    },
    {
      "amount": 1000000,
      "payment_method": "TRANSFER"
    }
  ]
}
```

### B. Pay Installment (Mendukung Split Repayment juga jika dibutuhkan)
`POST /api/v1/expenses/:id/payments`
```json
{
  "payments": [
    {
      "amount": 1500000,
      "payment_date": "2026-06-30T10:00:00Z",
      "payment_method": "CASH"
    }
  ]
}
```

---

## 6. Relasi Antar Tabel (ERD)

```
suppliers ──┐
             │ (optional)
             ▼
         expenses (Header Nota)
         ├──── expense_items (Detail Item, FK: expense_category_id → expense_categories)
         └──── expense_payments ──── cash_flows (1:1 mirror)
```

---

## 7. Panduan Eksekusi Mandiri (Malam Hari)

Gunakan panduan ini untuk melakukan migrasi database dan refactoring kode backend secara manual nanti malam.

### A. SQL Migration: Refactor Expenses & Create Items/Payments
Buat migration up (`000017_refactor_expenses_tables.up.sql`):
```sql
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
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS due_date TIMESTAMP;

-- 3. Buat tabel expense_items
CREATE TABLE IF NOT EXISTS expense_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    expense_category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    description VARCHAR(255),
    qty INT NOT NULL DEFAULT 1,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_expense_items_deleted_at ON expense_items(deleted_at);

-- 4. Buat tabel expense_payments
CREATE TABLE IF NOT EXISTS expense_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    payment_date TIMESTAMP NOT NULL DEFAULT NOW(),
    payment_method VARCHAR(50) NOT NULL,
    cashier_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_expense_payments_deleted_at ON expense_payments(deleted_at);

-- 5. Seed Kategori Khusus Diskon
INSERT INTO expense_categories (id, name, "group", created_at, updated_at)
VALUES (gen_random_uuid(), 'Potongan Pembelian', 'OPERATIONAL', NOW(), NOW())
ON CONFLICT DO NOTHING;
```

Dan migration down (`000017_refactor_expenses_tables.down.sql`):
```sql
DROP TABLE IF EXISTS expense_payments;
DROP TABLE IF EXISTS expense_items;
ALTER TABLE expenses DROP COLUMN IF EXISTS due_date;
ALTER TABLE expenses DROP COLUMN IF EXISTS status;
ALTER TABLE expenses DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE expenses DROP COLUMN IF EXISTS supplier_id;
ALTER TABLE expenses DROP COLUMN IF EXISTS invoice_number;
ALTER TABLE expenses DROP COLUMN IF EXISTS expense_number;
-- Kembalikan kolom lama jika diperlukan roll back penuh
```

---

### B. Update Model Struct di Go (`internal/expense/domain/expense.go`)

```go
type Expense struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseNumber string           `gorm:"type:varchar(100);uniqueIndex;not null" json:"expense_number"`
	InvoiceNumber *string          `gorm:"type:varchar(100)" json:"invoice_number"`
	SupplierID    *uuid.UUID       `gorm:"type:uuid" json:"supplier_id"`
	VendorName    string           `gorm:"type:varchar(255);not null" json:"vendor_name"`
	Amount        float64          `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status        string           `gorm:"type:varchar(50);not null;default:'PAID'" json:"status"`
	DueDate       *time.Time       `gorm:"type:timestamp" json:"due_date"`
	ExpenseDate   time.Time        `gorm:"type:timestamp;default:now()" json:"expense_date"`
	Description   *string          `gorm:"type:text" json:"description"`
	CashierID     uuid.UUID        `gorm:"type:uuid;not null" json:"cashier_id"`
	
	// Relations
	Items         []ExpenseItem    `gorm:"foreignKey:ExpenseID" json:"items"`
	Payments      []ExpensePayment `gorm:"foreignKey:ExpenseID" json:"payments,omitempty"`
	
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

type ExpenseItem struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseID         uuid.UUID       `gorm:"type:uuid;not null" json:"expense_id"`
	ExpenseCategoryID uuid.UUID       `gorm:"type:uuid;not null" json:"expense_category_id"`
	ExpenseCategory   ExpenseCategory `gorm:"foreignKey:ExpenseCategoryID" json:"expense_category"`
	Description       *string         `gorm:"type:varchar(255)" json:"description"`
	Qty               int             `gorm:"type:int;not null;default:1" json:"qty"`
	Price             float64         `gorm:"type:decimal(15,2);not null" json:"price"`
	Amount            float64         `gorm:"type:decimal(15,2);not null" json:"amount"` // Qty * Price
	
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

type ExpensePayment struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseID     uuid.UUID      `gorm:"type:uuid;not null" json:"expense_id"`
	Amount        float64        `gorm:"type:decimal(15,2);not null" json:"amount"`
	PaymentDate   time.Time      `gorm:"type:timestamp;default:now()" json:"payment_date"`
	PaymentMethod string         `gorm:"type:varchar(50);not null" json:"payment_method"`
	CashierID     uuid.UUID      `gorm:"type:uuid;not null" json:"cashier_id"`
	
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

---

### C. Alur Kode Transaksi GORM di Usecase (`expense_usecase.go`)

#### 1. Saat Create Expense:
```go
// Di dalam DB transaction (tx):
// 1. Generate ExpenseNumber (misal: format EXP/YYYYMMDD/XXXX)
// 2. Hitung sum(amount) dari detail items, set ke expense.Amount
// 3. Simpan header: tx.Create(&expense)
// 4. Simpan detail items: tx.Create(&expense.Items)
// 5. Jika status == "PAID" atau "PARTIAL" (ada initial_payment > 0):
//    - Simpan record di expense_payments: tx.Create(&payment)
//    - Lock & Deduct cash_accounts sesuai nominal bayar
//    - Simpan record cash_flows dengan reference_type = "EXPENSE_PAYMENT", reference_id = payment.ID
```

#### 2. Saat Bayar Cicilan:
```go
// Di dalam DB transaction (tx):
// 1. Lock & Get data `Expense` by ID
// 2. Hitung total yang sudah dibayar + cicilan baru
// 3. Pastikan total bayar <= expense.Amount
// 4. Simpan record di expense_payments: tx.Create(&newPayment)
// 5. Lock & Deduct cash_accounts
// 6. Simpan record cash_flows (ref_type: "EXPENSE_PAYMENT", ref_id: newPayment.ID)
// 7. Update status di `Expense`:
//    - Jika total bayar == expense.Amount -> set PAID
//    - Jika total bayar < expense.Amount -> set PARTIAL
//    - Simpan update: tx.Save(&expense)
```

