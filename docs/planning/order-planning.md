# Order & Job Entry Planning — ERP Digital Printing

Dokumen ini mendefinisikan arsitektur, siklus hidup transaksi, struktur database, serta alur kerja kolaboratif antara **Desainer (Job Entry)** dan **Kasir (POS & Billing)** untuk transaksi percetakan digital.

---

## 1. Pemisahan Tugas (Separation of Duties)

Operasional toko fisik membagi peran secara ketat untuk mencegah kecurangan (fraud) dan mengoptimalkan fokus kerja:

*   **Desainer (Job Entry)**:
    *   Fokus pada pemeriksaan kelayakan file cetak, ukuran (dimensi), bahan, finishing, dan kuantitas.
    *   **STRICT**: Desainer tidak melihat nominal harga cetak sama sekali untuk menghindari negosiasi langsung di meja depan.
    *   Data keranjang disimpan di **Local Storage** desainer sebelum di-submit ke database.
*   **Kasir (POS / Billing)**:
    *   Menerima antrean draf pesanan dari desainer.
    *   Melakukan kalkulasi harga otomatis menggunakan *Pricing Engine* berdasarkan jenis pelanggan (End User atau Reseller).
    *   Menerima pembayaran (DP/Lunas), menerbitkan nomor Invoice (`INV`), dan merilis SPK ke produksi.
*   **Operator Cetak (Production Workspace)**:
    *   Memproses antrean cetak sesuai dengan panduan teknis visual (SPGK) dari desainer tanpa mengetahui detail pembayaran.

---

## 2. Siklus Hidup & Transisi Status Order

Transisi status dikelola secara terpusat dalam satu tabel `orders` melalui status-status berikut:

| Status | Lokasi Data | Keterangan |
| :--- | :--- | :--- |
| **`DRAFT`** | Database | Tiket disimpan desainer ke database untuk diedit kembali nanti (misal: hold pesanan). Belum tayang di kasir. |
| **`PENDING_PAYMENT`** | Database | Tiket disubmit oleh desainer ke kasir. Kunci spesifikasi terkunci, masuk antrean POS kasir. |
| **`IN_PRODUCTION`** | Database | Kasir telah menerima pembayaran (DP atau Lunas). Nomor `invoice_number` diterbitkan dan SPK rilis ke workshop. |
| **`READY_FOR_PICKUP`** | Database | Operator telah menyelesaikan produksi dan lolos Quality Control (QC). Barang siap diserahkan. |
| **`COMPLETED`** | Database | Pelanggan mengambil barang dan sisa tagihan (jika ada) telah dilunasi 100% di kasir. |
| **`CANCELLED`** | Database | Transaksi dibatalkan oleh kasir/admin. Disimpan untuk kebutuhan audit keuangan. |

---

## 3. Skema Identitas Tiket vs Nota Invoice

Untuk menjaga kerapian pembukuan keuangan dan efisiensi pelacakan produksi, sistem membedakan nomor identitas:

*   **`job_number`** (Nomor Tiket):
    *   *Karakter*: Unique, Not Null.
    *   *Kapan dibuat*: Saat status pertama kali disubmit menjadi `DRAFT` atau `PENDING_PAYMENT` oleh desainer.
    *   *Format*: `JOB/YYYYMMDD/XXXX` (e.g. `JOB/20260530/0001`).
*   **`invoice_number`** (Nomor Nota Invoice):
    *   *Karakter*: Unique, Nullable.
    *   *Kapan dibuat*: Hanya ketika status berganti ke `IN_PRODUCTION` setelah kasir memproses pembayaran (DP atau Lunas).
    *   *Format*: `INV/YYYYMMDD/XXXX` (e.g. `INV/20260530/0001`).

---

## 4. Skema Hubungan Database & Aturan Bisnis

### A. Tabel Utama
*   **`orders`**: Menyimpan header transaksi, data historis denormalisasi pelanggan, nomor tiket/invoice, status, serta total biaya keseluruhan.
*   **`order_items`**: Menyimpan detail item cetakan (panjang, lebar, qty, file cetak, UOM, dan subtotal).
*   **`finishings`**: Master data untuk biaya tambahan opsi finishing percetakan (seperti mata ayam, laminasi, dll).
*   **`order_item_finishings`**: Pivot table Many-to-Many yang menghubungkan detail item dengan satu atau lebih finishing.

### B. Aturan Validasi UOM (Unit of Measurement)
Saat desainer melakukan submit draf tiket kerja:
1.  **Jika `uom == "m2"`**: Kolom `length_cm` dan `width_cm` **wajib** diisi bernilai positif (`> 0`).
2.  **Jika `uom == "m_lari"`**: Kolom `length_cm` **wajib** diisi bernilai positif (`> 0`). Lebar otomatis diabaikan karena sesuai lebar roll bahan.
3.  **Jika `uom == "box"` atau `"pcs"`**: Kolom `length_cm` dan `width_cm` otomatis di-set ke `NULL` di database.
