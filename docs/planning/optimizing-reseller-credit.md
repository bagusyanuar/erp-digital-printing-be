# Planning: Optimizing Reseller Credit & Receivables

Dokumen ini mencatat rencana jangka panjang untuk optimasi pengelolaan limit kredit dan kalkulasi total piutang (*outstanding debt*) per reseller pada sistem ERP Digital Printing.

---

## 1. Kondisi Saat Ini (Fase Awal)
* **Kalkulasi**: Menggunakan query agregasi `SUM(total_amount - paid_amount)` secara dinamis pada tabel `orders` yang difilter berdasarkan `reseller_id`, `payment_status` (`UNPAID`, `PARTIAL_PAID`), dan `status` (`!CANCELLED`).
* **Karakteristik**:
  * Real-time dan konsisten.
  * Cukup efisien untuk skala data kecil hingga menengah (dengan bantuan indexing).
  * Menjadi bottleneck ketika jumlah baris transaksi order mencapai ratusan ribu atau jutaan data.

---

## 2. Rencana Optimasi Jangka Panjang

Untuk mencegah degradasi performa di kemudian hari, kita merencanakan migrasi ke metode **State-based Tracking** (menyimpan saldo berjalan).

### Pendekatan: Kolom `credit_used` pada Tabel Reseller

Kita akan melacak penggunaan kredit secara langsung pada entitas Reseller (atau tabel mutasi saldo kredit khusus) alih-alih menghitung ulang dari tabel order setiap saat.

#### A. Perubahan Skema Database
Menambahkan field baru pada model `Reseller`:
```go
type Reseller struct {
    gorm.Model
    // ... field lainnya ...
    CreditLimit  int64 `gorm:"default:0"` // Batas maksimal kredit
    CreditUsed   int64 `gorm:"default:0"` // Total piutang berjalan (Outstanding Debt)
}
```

#### B. Mekanisme Sinkronisasi (Event-Driven / Hooks)
Setiap kali ada perubahan status transaksi atau pembayaran, kita melakukan mutasi terhadap nilai `CreditUsed`:

1. **Order Baru (Tempo)**:
   * Saat order dengan metode tempo dibuat dan disetujui.
   * Aksi: `CreditUsed = CreditUsed + (total_amount - paid_amount)`.
   * Validasi: Pastikan `CreditUsed + order_debt <= CreditLimit`.

2. **Pembayaran / Pelunasan (Payment)**:
   * Saat kasir mencatat pembayaran baru untuk order tempo tersebut.
   * Aksi: `CreditUsed = CreditUsed - payment_amount`.

3. **Pembatalan Order (Cancelled)**:
   * Saat order dibatalkan sebelum lunas.
   * Aksi: `CreditUsed = CreditUsed - sisa_piutang_order`.

4. **Retur / Penyesuaian Harga**:
   * Jika ada perubahan nilai order yang mengurangi tagihan.
   * Aksi: Menyesuaikan selisihnya pada `CreditUsed`.

---

## 3. Strategi Mitigasi Selisih Data (Reconciliation)

Kelemahan utama dari metode state-based tracking adalah risiko terjadinya selisih angka akibat *race condition* atau kegagalan sistem di tengah proses. Untuk memitigasi hal ini:

* **Database Transaction (ACID)**: Setiap update ke tabel `orders` dan `resellers` wajib dibungkus dalam satu transaksi database (`db.Transaction`).
* **Reconciliation Job**: Membuat background job mingguan/bulanan yang melakukan pengecekan ulang:
  * Sistem akan menghitung ulang `SUM(total_amount - paid_amount)` dari tabel `orders` secara *off-peak* (saat sepi pengunjung).
  * Jika ditemukan selisih dengan `CreditUsed` di tabel reseller, sistem akan memperbarui nilai `CreditUsed` dan mencatat log rekonsiliasi.

---

## 4. Keuntungan Setelah Optimasi
* **Kecepatan Validasi**: Proses pembuatan order tempo hanya membutuhkan pengecekan nilai kolom `CreditUsed` (O(1) complexity) tanpa membebani database dengan query agregasi tabel `orders`.
* **Performa List/Detail Reseller**: Endpoint rekap piutang hanya melakukan `SELECT` sederhana pada tabel reseller.
