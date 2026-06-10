# Konsep Arus Kas (Cash Flow / General Ledger) — ERP Digital Printing

Dokumen ini mendefinisikan rancangan sistem pencatatan dan pelaporan arus kas (cash flow) menggunakan pendekatan **General Ledger (Buku Kas)**. Pendekatan ini memusatkan seluruh transaksi keuangan ke dalam satu tabel sebagai *Single Source of Truth*, mencatat pergerakan Debit (uang masuk) dan Credit (uang keluar).

---

## 1. Komponen Arus Kas

Sistem mencatat seluruh mutasi kas secara tersentralisasi pada tabel `cash_flows`. Setiap kejadian finansial di sistem harus melakukan penambahan *entry* ke tabel ini secara *atomic* (dalam satu transaksi database yang sama dengan operasi utamanya).

1.  **Arus Kas Masuk (DEBIT)**:
    *   `ORDER_PAYMENT`: Pembayaran pesanan lunas di depan, DP, atau pelunasan sisa piutang.
    *   `CAPITAL_INJECTION`: Penambahan modal atau setoran kasir.
    *   `ADJUSTMENT`: Penyesuaian kas (selisih lebih/surplus) saat Cash Opname.
2.  **Arus Kas Keluar (CREDIT)**:
    *   `EXPENSE`: Pengeluaran untuk bahan baku, operasional, listrik, kas kecil, dll.
    *   `REFUND`: Pengembalian dana ke pelanggan karena pembatalan/kesalahan order.
    *   `ADJUSTMENT`: Penyesuaian kas (selisih kurang/shortage) saat Cash Opname.
    *   `CAPITAL_WITHDRAWAL`: Penarikan dana oleh owner (Prive) dari kas/rekening perusahaan.

---

## 2. Model Data & Skema Database

Sistem membutuhkan tabel utama `cash_flows` dan tabel referensi `expenses` (untuk menyimpan detail pengeluaran sebelum masuk ke buku kas utama).

### A. Tabel Buku Kas Utama (`cash_flows`)
```dbml
Table cash_flows {
  id uuid [pk]
  transaction_date timestamp       // Waktu terjadinya mutasi
  reference_type varchar(50)       // 'ORDER_PAYMENT', 'EXPENSE', 'REFUND', 'CAPITAL', 'ADJUSTMENT'
  reference_id uuid                // ID referensi (bisa null jika penyesuaian manual tanpa tabel detail)
  type varchar(10)                 // 'DEBIT' (Masuk), 'CREDIT' (Keluar)
  amount decimal(15,2)
  payment_method varchar(50)       // 'cash', 'transfer', 'qris'
  description text                 // Keterangan otomatis (misal: "Pembayaran Invoice INV/2026/001") atau manual
  cashier_id uuid [ref: > users.id] // User/Kasir yang melakukan transaksi
  created_at timestamp
  updated_at timestamp
  deleted_at timestamp [index]
}
```

### B. Tabel Detail Pengeluaran (`expenses`)
Untuk mencatat detail nota/kategori pengeluaran sebelum jurnalnya dibuat.
```dbml
Table expenses {
  id uuid [pk]
  cashier_id uuid [ref: > users.id]
  amount decimal(15,2)
  category varchar(100)            // 'raw_material', 'operational', 'salary', 'utility', 'other'
  description text
  created_at timestamp
  updated_at timestamp
  deleted_at timestamp [index]
}
```
*Catatan: Setiap kali `expenses` dibuat, sistem wajib membuat record di `cash_flows` dengan `type='CREDIT'`, `reference_type='EXPENSE'`, `reference_id=expenses.id`.*

---

## 3. Spesifikasi API Laporan Arus Kas (Cash Flow Report)

Laporan sangat ringan karena hanya melakukan agregasi (SUM) pada satu tabel `cash_flows`.

### Get Cash Flow Summary & List
*   **Method & URL**: `GET /api/v1/reports/cash-flow`
*   **Query Parameters**:
    *   `start_date` (string, required, YYYY-MM-DD)
    *   `end_date` (string, required, YYYY-MM-DD)
*   **Response (JSON)**:
    ```json
    {
      "code": 200,
      "status": "OK",
      "message": "Cash flow report fetched successfully",
      "data": {
        "summary": {
          "total_debit": 15000000.00,
          "total_credit": 5000000.00,
          "net_balance": 10000000.00
        },
        "details_by_method": {
          "cash": { "debit": 5000000, "credit": 2000000, "balance": 3000000 },
          "transfer": { "debit": 8000000, "credit": 3000000, "balance": 5000000 },
          "qris": { "debit": 2000000, "credit": 0, "balance": 2000000 }
        },
        "transactions": [
          {
            "id": "uuid",
            "transaction_date": "2026-06-11T10:00:00Z",
            "reference_type": "ORDER_PAYMENT",
            "type": "DEBIT",
            "amount": 500000.00,
            "payment_method": "cash",
            "description": "Pelunasan INV/20260611/0001",
            "cashier_name": "Kasir Pagi"
          }
        ]
      }
    }
    ```

---

## 4. API Pengelolaan Pengeluaran (Expenses API)

1.  **Tambah Catatan Pengeluaran**:
    *   `POST /api/v1/expenses`
    *   Request Body:
        ```json
        {
          "amount": 250000,
          "category": "operational",
          "payment_method": "cash",
          "description": "Pembelian token listrik workshop"
        }
        ```
    *   *Sistem akan otomatis meng-insert juga ke tabel `cash_flows` dengan tipe `CREDIT`.*
2.  **Daftar Pengeluaran (Paginated)**:
    *   `GET /api/v1/expenses?page=1&limit=10&search=listrik`
3.  **Hapus Catatan Pengeluaran**:
    *   `DELETE /api/v1/expenses/:id`
    *   *Penghapusan (Soft Delete) ini juga wajib men-soft-delete entri yang bersesuaian di tabel `cash_flows`.*

---

## 5. Penanganan Kasus Khusus (Edge Cases)

Pendekatan *General Ledger* secara *native* dan elegan menangani berbagai kasus transaksi harian toko:

1.  **Split Payment (Pembayaran Kombinasi)**
    *   **Kasus**: Customer membayar total tagihan Rp 3.000 dengan rincian: Rp 1.000 tunai (Cash) dan Rp 2.000 QRIS.
    *   **Perilaku Sistem**: Sistem membuat 2 baris terpisah di tabel `order_payments`. *Trigger/Hook* akan langsung menerjemahkannya menjadi **2 baris mutasi** di tabel `cash_flows`:
        *   Baris 1: `DEBIT`, Rp 1.000, `payment_method = cash`
        *   Baris 2: `DEBIT`, Rp 2.000, `payment_method = qris`
    *   **Dampak**: Saldo laci kasir (fisik) dan saldo rekening digital dijamin tetap akurat tanpa bercampur.

2.  **Transaksi Piutang (Tempo / Hutang)**
    *   **Kasus**: Customer berhutang Rp 8.000 (belum dibayar sama sekali) atau baru membayar DP Rp 2.000 dari total Rp 10.000.
    *   **Perilaku Sistem**: Tabel `cash_flows` menerapkan prinsip **Cash Basis**. Nilai piutang Rp 8.000 **tidak akan dicatat** di `cash_flows` karena uang fisik belum diterima. Hanya DP Rp 2.000 yang masuk sebagai `DEBIT`.
    *   **Saat Pelunasan**: Ketika customer melunasi hutangnya beberapa hari kemudian, barulah sistem membuat entri baru di `cash_flows` sebesar Rp 8.000 (Tipe: `DEBIT`, Reference: `ORDER_PAYMENT` / Repayment).
    *   **Dampak**: Rekap kasir harian akan selalu cocok dengan uang fisik di laci karena nilai hutang dipisahkan ke modul *Accounts Receivable* (Piutang) dan tidak masuk ke Buku Kas (Ledger).
