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
* `subtotal` (DECIMAL(15,2)) — Total harga item setelah diskon per item (`SUM(expense_items.amount)`).
* `discount` (DECIMAL(15,2), Default 0) — Diskon tambahan di level nota.
* `amount` (DECIMAL(15,2)) — Total tagihan akhir (`subtotal - discount`). Ini yang jadi acuan pembayaran.
* `status` (VARCHAR) — `'PAID'`, `'PARTIAL'`, `'UNPAID'`.
* `due_date` (TIMESTAMP, Nullable) — Jatuh tempo, wajib diisi jika `PARTIAL` atau `UNPAID`.
* `expense_date` (TIMESTAMP) — Tanggal transaksi.
* `description` (TEXT, Nullable) — Catatan tambahan.
* `cashier_id` (UUID, Not Null) — ID kasir yang menginput.
* `created_at`, `updated_at`, `deleted_at`

---

### C. Tabel `expense_items` (Detail Item Per Nota)
Mencatat setiap baris item pengeluaran dalam 1 nota. Setiap item punya kategori sendiri (produksi/operasional).
* `id` (UUID, PK)
* `expense_id` (UUID, FK → expenses, Not Null)
* `expense_category_id` (UUID, FK → expense_categories, Not Null) — Kategori pengeluaran per item.
* `description` (VARCHAR, Nullable) — Keterangan item (e.g., "Kertas Art Paper 100gsm").
* `qty` (INTEGER, Default 1) — Jumlah.
* `price` (DECIMAL(15,2)) — Harga satuan.
* `discount` (DECIMAL(15,2), Default 0) — Diskon per item.
* `amount` (DECIMAL(15,2)) — Total per item (`qty × price - discount`).
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

## 2. Kalkulasi Diskon

Diskon tersedia di 2 level:

```
expense_items:
  Item 1: qty(10) × price(300.000) = 3.000.000 - discount(100.000) = 2.900.000
  Item 2: qty(5)  × price(300.000) = 1.500.000 - discount(0)       = 1.500.000
  Item 3: qty(1)  × price(500.000) = 500.000   - discount(0)       = 500.000

expenses:
  subtotal      = SUM(items.amount) = 4.900.000
  discount      = 200.000  (diskon nota dari supplier)
  amount        = 4.900.000 - 200.000 = 4.700.000  ← total tagihan
```

---

## 3. Alur Pengisian (Workflow)

### A. Input Vendor/Supplier
1. **Supplier Terdaftar**: Pilih dropdown → `supplier_id` = UUID, `vendor_name` = nama snapshot.
2. **Non-Supplier**: Ketik manual → `supplier_id` = NULL, `vendor_name` = teks input.

### B. Input Pembayaran
1. **PAID (Lunas)**: Otomatis buat 1 record `expense_payments` senilai `expenses.amount`.
2. **UNPAID (Hutang Penuh)**: `due_date` wajib diisi. Tidak ada record `expense_payments`.
3. **PARTIAL (Termin/DP)**: `due_date` wajib diisi. Buat 1 record `expense_payments` senilai DP.

### C. Pembayaran Cicilan Berikutnya
1. Jalankan DB Transaction (`tx`).
2. Hitung total bayar: `SUM(expense_payments.amount) + amount_baru`.
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
Nota: Rp 5.000.000 (gross) - Diskon Rp 200.000 = Rp 4.800.000

Payment 1 (DP):
  expense_payments → amount: 2.000.000
  cash_flows       → CREDIT 2.000.000, ref_type: 'EXPENSE_PAYMENT'

Payment 2 (Pelunasan):
  expense_payments → amount: 2.800.000
  cash_flows       → CREDIT 2.800.000, ref_type: 'EXPENSE_PAYMENT'

Total uang keluar = 4.800.000 ✅
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
  "discount": 200000,
  "status": "PARTIAL",
  "due_date": "2026-07-16T00:00:00Z",
  "expense_date": "2026-06-16T12:00:00Z",
  "description": "Pembelian bahan cetak termin 1 bulan",
  "items": [
    {
      "expense_category_id": "uuid-kategori-produksi",
      "description": "Kertas Art Paper 100gsm",
      "qty": 10,
      "price": 300000,
      "discount": 100000
    },
    {
      "expense_category_id": "uuid-kategori-produksi",
      "description": "Tinta DTF",
      "qty": 5,
      "price": 300000,
      "discount": 0
    },
    {
      "expense_category_id": "uuid-kategori-operasional",
      "description": "Ongkos Kirim",
      "qty": 1,
      "price": 500000,
      "discount": 0
    }
  ],
  "initial_payment": 2000000,
  "payment_method": "TRANSFER"
}
```

### B. Pay Installment
`POST /api/v1/expenses/:id/payments`
```json
{
  "amount": 1500000,
  "payment_date": "2026-06-30T10:00:00Z",
  "payment_method": "CASH"
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
