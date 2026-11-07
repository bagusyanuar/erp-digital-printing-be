# 🛠️ Panduan Perbaikan Data: Kelebihan Catat Nominal Bayar (Overpayment & Kembalian Kasir)

Dokumen ini mencatat panduan investigasi dan script SQL perbaikan data akibat kasir memasukkan nominal uang fisik tunai yang diterima pelanggan (misal: bayar Rp 100.000 untuk tagihan Rp 70.000 dengan kembalian Rp 30.000), di mana sebelum perbaikan sistem mencatat nominal kotor tersebut ke dalam database.

---

## 📌 1. Latar Belakang Masalah

1. **Bug Alur Pembayaran**:
   * Kasir mengetik uang tunai yang diterima pelanggan di form pembayaran kasir.
   * Frontend sebelumnya mengirimkan nominal kotor tersebut mentah-mentah ke backend via payload `payments[].amount_paid`.
   * Backend menyimpan nilai tersebut ke:
     * `orders.amount_paid` (melebihi `grand_total`).
     * `order_payments.amount`.
     * `cash_flows.amount`.
     * `cash_accounts.balance` (saldo bertambah sebesar uang kotor, padahal uang fisik kembalian sudah diserahkan ke pelanggan).
2. **Dampak ke Laporan**:
   * Saldo kas di database menggelembung (*surplus semu*) dibanding fisik uang di laci kasir.
   * Modul piutang menghasilkan selisih minus (`grand_total - amount_paid < 0`).

---

## 🔍 2. Query Pengecekan Data Terdampak

Jalankan query ini di DBeaver untuk memeriksa transaksi yang terdampak:

### A. Rekap Total Saldo Kas/Bank yang Menggelembung
```sql
SELECT 
    op.payment_method,
    COUNT(DISTINCT o.id) AS total_order_terdampak,
    SUM(o.amount_paid - o.grand_total) AS total_saldo_gelembung
FROM orders o
JOIN order_payments op ON op.order_id = o.id AND op.deleted_at IS NULL
WHERE o.amount_paid > (o.grand_total + 0.01)
  AND o.deleted_at IS NULL
GROUP BY op.payment_method;
```

### B. Daftar Rinci Seluruh Order yang Mengalami Kelebihan Catat
```sql
SELECT 
    o.id,
    o.invoice_number,
    o.customer_name,
    o.payment_status,
    o.grand_total AS tagihan_asli,
    o.amount_paid AS uang_diterima_kasir,
    (o.amount_paid - o.grand_total) AS selisih_kembalian_gelembung,
    op.payment_method,
    o.created_at
FROM orders o
JOIN order_payments op ON op.order_id = o.id AND op.deleted_at IS NULL
WHERE o.amount_paid > (o.grand_total + 0.01)
  AND o.deleted_at IS NULL
ORDER BY o.created_at DESC;
```

---

## 🚀 3. Eksekusi Script SQL (DBeaver / `psql`)

> **Note**: Gunakan **Execute SQL Script** (`Option + X` di macOS / `Alt + X` di Windows) agar blok transaksi `BEGIN ... COMMIT` berjalan utuh.

### Skenario A: Batch Fixing Semua Transaksi Terdampak (Rekomendasi)
Script ini otomatis menormalkan seluruh order yang mengalami kelebihan bayar, menyelaraskan log mutasi kas, dan memotong saldo di `cash_accounts` sesuai total selisihnya:

```sql
BEGIN;

-- 1. Tampung transaksi terdampak ke temporary table
CREATE TEMP TABLE temp_affected_orders ON COMMIT DROP AS
SELECT 
    o.id AS order_id,
    o.grand_total,
    (o.amount_paid - o.grand_total) AS diff,
    op.id AS payment_id,
    op.payment_method
FROM orders o
JOIN order_payments op ON op.order_id = o.id AND op.deleted_at IS NULL
WHERE o.amount_paid > (o.grand_total + 0.01)
  AND o.deleted_at IS NULL;

-- 2. Normalkan amount_paid di tabel orders (disamakan ke grand_total)
UPDATE orders o
SET 
    amount_paid = o.grand_total,
    payment_status = 'PAID',
    updated_at = NOW()
FROM temp_affected_orders t
WHERE o.id = t.order_id;

-- 3. Normalkan nominal di order_payments
UPDATE order_payments op
SET 
    amount = op.amount - t.diff,
    updated_at = NOW()
FROM temp_affected_orders t
WHERE op.id = t.payment_id;

-- 4. Normalkan nominal mutasi di cash_flows
UPDATE cash_flows cf
SET 
    amount = cf.amount - t.diff,
    updated_at = NOW()
FROM temp_affected_orders t
WHERE cf.reference_id = t.payment_id
  AND cf.reference_type = 'ORDER_PAYMENT'
  AND cf.deleted_at IS NULL;

-- 5. Potong saldo menggelembung di cash_accounts sesuai total per payment method
UPDATE cash_accounts ca
SET 
    balance = ca.balance - sub.total_diff,
    updated_at = NOW()
FROM (
    SELECT payment_method, SUM(diff) AS total_diff
    FROM temp_affected_orders
    GROUP BY payment_method
) sub
WHERE LOWER(ca.name) = LOWER(sub.payment_method);

COMMIT;
```

---

### Skenario B: Single Fixing untuk 1 Invoice Tertentu (Custom Harga)
Gunakan jika hanya ingin memperbaiki 1 invoice spesifik (contoh: `INV/202611/00048` dengan penyesuaian harga satuan Rp 5.000):

```sql
BEGIN;

WITH params AS (
    SELECT 
        'INV/202611/00048'::text AS target_invoice, -- Ganti nomor invoice
        5000::numeric            AS new_price_per_unit -- Ganti harga satuan baru
),
target_order AS (
    SELECT o.id AS order_id, o.invoice_number
    FROM orders o
    JOIN params p ON o.invoice_number = p.target_invoice
),
old_payment_info AS (
    SELECT 
        op.id AS payment_id,
        op.payment_method,
        op.amount AS old_amount
    FROM order_payments op
    JOIN target_order t ON op.order_id = t.order_id
    WHERE op.deleted_at IS NULL
    LIMIT 1
),
updated_items AS (
    UPDATE order_items oi
    SET 
        price_per_unit = p.new_price_per_unit,
        subtotal = ROUND(
            (
                p.new_price_per_unit * CASE 
                    WHEN oi.uom = 'm2' THEN (COALESCE(oi.length_m, 1) * COALESCE(oi.width_m, 1) * oi.quantity)
                    WHEN oi.uom = 'm_lari' THEN (COALESCE(oi.length_m, 1) * oi.quantity)
                    ELSE oi.quantity
                END
            ) + (COALESCE(oi.additional_cost, 0) * oi.quantity),
            2
        ),
        updated_at = NOW()
    FROM target_order t, params p
    WHERE oi.order_id = t.order_id
      AND oi.deleted_at IS NULL
    RETURNING oi.order_id, oi.subtotal, oi.additional_cost, oi.quantity
),
aggregated AS (
    SELECT 
        order_id,
        SUM(subtotal - (COALESCE(additional_cost, 0) * quantity)) AS new_product_price,
        SUM(COALESCE(additional_cost, 0) * quantity) AS new_additional_cost,
        SUM(subtotal) AS new_grand_total
    FROM updated_items
    GROUP BY order_id
),
updated_order AS (
    UPDATE orders o
    SET 
        total_product_price = a.new_product_price,
        total_additional_cost = a.new_additional_cost,
        grand_total = a.new_grand_total,
        amount_paid = a.new_grand_total,
        payment_status = 'PAID',
        updated_at = NOW()
    FROM aggregated a
    WHERE o.id = a.order_id
    RETURNING o.id, o.invoice_number, o.grand_total AS final_amount
),
updated_payments AS (
    UPDATE order_payments op
    SET 
        amount = uo.final_amount,
        payment_type = 'FULL_PAYMENT',
        updated_at = NOW()
    FROM updated_order uo, old_payment_info old
    WHERE op.id = old.payment_id
),
updated_cash_flows AS (
    UPDATE cash_flows cf
    SET 
        amount = uo.final_amount,
        updated_at = NOW()
    FROM updated_order uo
    WHERE cf.invoice_number = uo.invoice_number
      AND cf.reference_type = 'ORDER_PAYMENT'
      AND cf.deleted_at IS NULL
)
UPDATE cash_accounts ca
SET 
    balance = ca.balance - (old.old_amount - uo.final_amount),
    updated_at = NOW()
FROM old_payment_info old, updated_order uo
WHERE LOWER(ca.name) = LOWER(old.payment_method);

COMMIT;
```

---

### Skenario C: Single Fixing Invoice Belum Lunas (DP / Tempo / Custom Item Tertentu)
Gunakan jika invoice **belum lunas** dan hanya item tertentu yang harganya di-override (misal `AC260 2 sisi = 4500`), sedangkan item lainnya tetap menggunakan harga master tier sistem:
* `orders.amount_paid`, `order_payments`, `cash_flows`, dan `cash_accounts` **TIDAK DIUBAH** (karena uang DP/cicilan yang masuk adalah riil).
* Grand total dan sisa piutang otomatis mengecil mengikuti harga baru.

```sql
ROLLBACK; -- Reset jika sesi sebelumnya sempat error

BEGIN;

WITH target_order AS (
    SELECT 
        o.id AS order_id, 
        COALESCE(r.customer_level_id, 'b3c8f3a3-b26a-4638-b7f2-841a54774844'::uuid) AS customer_level_id
    FROM orders o
    LEFT JOIN resellers r ON r.id = o.reseller_id
    WHERE o.invoice_number = 'INV/202611/XXXXX' -- Ganti nomor invoice
),
-- 1. Hitung harga target di SELECT CTE (khusus item tertentu di-override, sisanya harga master)
calc_items AS (
    SELECT 
        oi.id AS item_id,
        CASE 
            WHEN (
                (pv.variant_name ILIKE '%AC260%' OR p.name ILIKE '%AC260%') 
                AND (pv.variant_name ILIKE '%2%sis%' OR pv.variant_name ILIKE '%2sisi%' OR p.name ILIKE '%2%sis%')
            ) THEN 4500 -- Ganti nominal harga khusus item target
            ELSE COALESCE(pt.price_per_unit, oi.price_per_unit) + COALESCE(pv.additional_cost, 0)
        END AS target_unit_price,
        CASE 
            WHEN oi.uom = 'm2' THEN (COALESCE(oi.length_m, 1) * COALESCE(oi.width_m, 1) * oi.quantity)
            WHEN oi.uom = 'm_lari' THEN (COALESCE(oi.length_m, 1) * oi.quantity)
            ELSE oi.quantity
        END AS calc_qty,
        COALESCE(oi.additional_cost, 0) AS finishing_cost,
        oi.quantity
    FROM order_items oi
    JOIN target_order t ON oi.order_id = t.order_id
    JOIN product_variants pv ON pv.id = oi.product_variant_id
    JOIN products p ON p.id = pv.product_id
    LEFT JOIN price_tiers pt ON pt.product_variant_id = pv.id
        AND pt.customer_level_id = t.customer_level_id
        AND pt.min_qty <= oi.quantity
        AND (pt.max_qty >= oi.quantity OR pt.max_qty IS NULL)
    WHERE oi.deleted_at IS NULL
),
-- 2. Update order_items
updated_items AS (
    UPDATE order_items oi
    SET 
        price_per_unit = c.target_unit_price,
        subtotal = ROUND((c.target_unit_price * c.calc_qty) + (c.finishing_cost * c.quantity), 2),
        updated_at = NOW()
    FROM calc_items c
    WHERE oi.id = c.item_id
    RETURNING oi.order_id, oi.subtotal, oi.additional_cost, oi.quantity
),
-- 3. Hitung grand_total baru
aggregated AS (
    SELECT 
        order_id,
        SUM(subtotal - (COALESCE(additional_cost, 0) * quantity)) AS new_product_price,
        SUM(COALESCE(additional_cost, 0) * quantity) AS new_additional_cost,
        SUM(subtotal) AS new_grand_total
    FROM updated_items
    GROUP BY order_id
)
-- 4. Update orders (amount_paid TETAP, sisa piutang otomatis sinkron)
UPDATE orders o
SET 
    total_product_price = a.new_product_price,
    total_additional_cost = a.new_additional_cost,
    grand_total = a.new_grand_total,
    payment_status = CASE 
        WHEN o.amount_paid >= (a.new_grand_total - 0.01) THEN 'PAID'
        WHEN o.amount_paid > 0 THEN 'PARTIAL_PAID'
        ELSE 'UNPAID'
    END,
    updated_at = NOW()
FROM aggregated a
WHERE o.id = a.order_id;

COMMIT;
```

---

## ✅ 4. Verifikasi Pasca Eksekusi

Jalankan query ini untuk memastikan semua saldo dan order sudah bersih dan sinkron:

```sql
-- 1. Verifikasi tidak ada lagi order overpaid (Hasil harus 0 rows)
SELECT 
    COUNT(*) AS total_overpaid_orders
FROM orders 
WHERE amount_paid > (grand_total + 0.01) 
  AND deleted_at IS NULL;

-- 2. Cek saldo akun kas/bank terkini
SELECT id, name, balance, updated_at 
FROM cash_accounts;
```

