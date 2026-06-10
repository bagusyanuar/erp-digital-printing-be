# Order & Job Entry Planning V2 — ERP Digital Printing

Dokumen ini mendefinisikan arsitektur pembaruan alur transaksi draf pesanan, di mana kalkulasi harga (Pricing Engine) dilakukan langsung saat Designer membuat draf (`DRAFT`) atau men-submit tiket (`PENDING_PAYMENT`).

---

## 1. Pemisahan Peran & Perubahan Alur (Workflow Changes)

Pada versi sebelumnya, harga baru dihitung saat pembayaran di kasir. Pada **V2**, alur ditingkatkan agar Designer dapat langsung melihat/menyimpan harga draf dengan aturan berikut:

*   **Designer (Job Entry)**:
    *   Menginput spesifikasi produk cetak (dimensi, bahan, qty, finishing).
    *   **Wajib** menentukan profil harga pelanggan (`customer_level_id` atau `reseller_id`) di awal.
    *   Sistem langsung menghitung harga secara real-time saat draf disimpan ke database, sehingga nominal harga di draf sudah valid dan akurat.
*   **Kasir (POS / Billing)**:
    *   Menerima antrean order (`PENDING_PAYMENT`) yang sudah memiliki kalkulasi harga lengkap.
    *   Kasir hanya melakukan konfirmasi pembayaran, penyesuaian diskon manual (jika ada), menerbitkan nomor `invoice_number`, dan mengubah status ke `IN_PRODUCTION`.

---

## 2. Siklus Hidup & Transisi Status Order (V2)

### A. Status Alur Kerja Produksi (`status`)
| Status | Lokasi Data | Keterangan |
| :--- | :--- | :--- |
| **`DRAFT`** | Database | Draf pesanan disimpan desainer. **Harga sudah terkalkulasi**. Belum tayang di kasir. |
| **`PENDING_PAYMENT`** | Database | Tiket masuk antrean kasir. Spesifikasi & harga dikunci. |
| **`IN_PRODUCTION`** | Database | Pembayaran diproses kasir (lunas, DP/partial, atau tempo). Nomor `invoice_number` terbit, SPK rilis ke workshop. |
| **`READY_FOR_PICKUP`** | Database | Produksi selesai & lolos Quality Control (QC). |
| **`COMPLETED`** | Database | Barang diambil & sisa tagihan dilunasi 100% (payment_status wajib `PAID`). |
| **`CANCELLED`** | Database | Transaksi dibatalkan. |

### B. Status Keuangan (`payment_status`)
*   **`UNPAID`**: Belum ada pembayaran masuk sama sekali (termasuk skema Tempo / DP = 0).
*   **`PARTIAL_PAID`**: Pembayaran uang muka (DP) > 0 tetapi kurang dari `grand_total`. Sisa pembayaran menjadi piutang.
*   **`PAID`**: Transaksi telah dilunasi 100%.

### C. Matriks Hubungan Status Produksi & Pembayaran
| Status Produksi (`status`) | Status Keuangan (`payment_status`) | Nominal Bayar (`amount_paid`) vs Tagihan (`grand_total`) | Keterangan Bisnis |
| :--- | :--- | :--- | :--- |
| **`DRAFT`** | `UNPAID` | Belum diinput | Draf desainer. |
| **`PENDING_PAYMENT`** | `UNPAID` | Belum diinput | Antrean di kasir POS. |
| **`IN_PRODUCTION`** | `PAID` | `amount_paid >= grand_total` | Lunas di depan langsung diproduksi. |
| **`IN_PRODUCTION`** | `PARTIAL_PAID` | `0 < amount_paid < grand_total` | Bayar DP sebagian, masuk produksi. Sisa jadi piutang. |
| **`IN_PRODUCTION`** | `UNPAID` | `amount_paid == 0` | Khusus Reseller/Tempo. Naik cetak tanpa DP (validasi limit kredit). |
| **`READY_FOR_PICKUP`** | `PAID` / `PARTIAL_PAID` / `UNPAID` | Mengikuti kondisi terakhir | Barang selesai cetak, siap diambil. |
| **`COMPLETED`** | `PAID` | Wajib `amount_paid >= grand_total` | Pelunasan saat barang diambil (serah terima barang). |

---

## 3. Penambahan Field Payload pada Save Draft (V2)

Untuk mendukung kalkulasi harga instan, payload request `Save Draft` & `Submit to Cashier` dilengkapi dengan parameter reseller dan data pelanggan:

*   **`reseller_id`** (UUID, Optional): Jika diisi, sistem akan otomatis mendeteksi level reseller tersebut untuk mengambil tier harga reseller.
*   **`customer_name`** (string, Optional): Nama pelanggan.
*   **`customer_phone`** (string, Optional): Nomor telepon pelanggan.

---

## 4. Spesifikasi API Endpoint Draf V2

### Save Job Entry Draft (Designer Save)
*   **Method & URL**: `POST /api/v1/orders/draft`
*   **Request Body (JSON)**:
    ```json
    {
      "designer_id": "uuid (required)",
      "reseller_id": "uuid (optional)",
      "customer_name": "string (optional)",
      "customer_phone": "string (optional)",
      "notes": "string (optional)",
      "items": [
        {
          "product_variant_id": "uuid (required)",
          "uom": "string (required, oneof=pcs m2 m_lari box)",
          "length_cm": "float64 (optional)",
          "width_cm": "float64 (optional)",
          "quantity": "int (required, >0)",
          "design_file_url": "string (optional)",
          "production_notes": "string (optional)",
          "finishing_ids": ["uuid"]
        }
      ]
    }
    ```

*   **Response (JSON)**:
    ```json
    {
      "code": 201,
      "status": "Created",
      "message": "Draft order saved successfully",
      "data": {
        "id": "uuid",
        "job_number": "JOB/20260531/0001",
        "invoice_number": null,
        "reseller_id": "uuid/null",
        "designer_id": "uuid",
        "status": "DRAFT",
        "payment_status": "UNPAID",
        "total_product_price": 150000.00,
        "total_additional_cost": 20000.00,
        "grand_total": 170000.00,
        "order_items": [
          {
            "id": "uuid",
            "product_variant_id": "uuid",
            "uom": "m2",
            "length_cm": 200,
            "width_cm": 100,
            "quantity": 1,
            "price_per_unit": 75000.00,
            "additional_cost": 20000.00,
            "subtotal": 170000.00,
            "finishings": [
              {
                "id": "uuid",
                "name": "Mata Ayam",
                "price": 5000.00
              }
            ]
          }
        ]
      }
    }
    ```

---

## 5. Validasi Bisnis & Kalkulasi di Backend

1.  **Resolusi Level Pelanggan**:
    *   Jika `reseller_id` dikirim dan valid, ambil `customer_level_id` dari tabel `resellers`.
    *   Jika tidak, gunakan `customer_level_id` yang dikirim dari payload.
    *   Jika keduanya kosong, gunakan default level **End User** (`b3c8f3a3-b26a-4638-b7f2-841a54774844`).
2.  **Validasi Kuantitas & Dimensi**:
    *   Validasi UOM (`m2`, `m_lari`, `pcs`, `box`) tetap dijalankan ketat.
    *   Kalkulasi kuantitas riil/luas cetak (panjang x lebar x qty) dilakukan sebelum pencarian harga ke pricing engine.
3.  **Matriks Harga**:
    *   Harga satuan per item cetak dicari dari tabel `price_tiers` menggunakan `CheckPrice`.
    *   Subtotal item = `(price_per_unit * calculated_qty) + (finishing_cost * order_quantity)`.

---

## 6. Fitur Manual Price Override oleh Kasir

Sesuai arahan Owner, Kasir memiliki wewenang penuh untuk melakukan **Manual Price Override** (penimpaan harga satuan secara manual) saat memproses transaksi pembayaran di POS.

### Skema Endpoint Pembayaran POS V2
*   **Method & URL**: `POST /api/v1/orders/:id/pay`
*   **Request Body (JSON)**:
    ```json
    {
      "reseller_id": "uuid (optional)",
      "customer_name": "string (required)",
      "customer_phone": "string (required)",
      "payment_method": "string (required, e.g. cash, transfer, tempo)",
      "payment_type": "string (required, oneof=full tempo)",
      "amount_paid": "float64 (required, >=0)",
      "price_overrides": [
        {
          "order_item_id": "uuid (required)",
          "price_per_unit": "float64 (required, >=0)"
        }
      ]
    }
    ```

### Aturan Bisnis Override Harga
1.  Jika array `price_overrides` dikirim oleh Kasir, sistem akan mencocokkan `order_item_id` dan **menimpa** `price_per_unit` bawaan pricing engine dengan nilai kustom tersebut.
2.  Jika tidak ada override untuk item tertentu, sistem tetap menggunakan harga standar dari matriks `price_tiers`.
3.  Subtotal item dan Grand Total order akan dihitung ulang secara real-time berdasarkan harga override tersebut sebelum menyimpan status pembayaran (`IN_PRODUCTION`).

---

## 7. Transaksi Hutang (Tempo) & Validasi Credit Limit Reseller

Khusus untuk pelanggan bertipe **Reseller**, sistem mendukung metode pembayaran kredit / hutang jangka pendek (**Tempo**). Pembayaran tempo diatur ketat dengan **Credit Limit** yang dimiliki oleh reseller.

### A. Perubahan Skema Request POS Pembayaran (`POST /api/v1/orders/:id/pay`)
*   `payment_method`: Menambahkan nilai `"tempo"` selain `"cash"`, `"transfer"`, atau `"qris"`.
*   `payment_type`: Menambahkan nilai `"tempo"` (selain `"full"`).
*   `amount_paid`: Boleh di-set `0` (atau lebih jika membayar DP parsial dari total tagihan tempo).

### B. Alur Kerja Validasi Credit Limit di Usecase
1.  **Deteksi Transaksi Tempo**:
    *   Jika `payment_method == "tempo"` atau `payment_type == "tempo"`.
2.  **Kalkulasi Outstanding Debt**:
    *   Sistem menghitung total hutang berjalan dari reseller tersebut:
        $$\text{Outstanding Debt} = \sum (\text{grand\_total} - \text{amount\_paid})$$
        untuk semua order milik `reseller_id` dengan status pembayaran `UNPAID` atau `PARTIAL_PAID`.
3.  **Validasi Limit**:
    *   Sistem menghitung perkiraan hutang baru jika transaksi ini disetujui:
        $$\text{New Potential Debt} = \text{Outstanding Debt} + (\text{Grand Total Order Baru} - \text{Amount Paid Order Baru})$$
    *   **STRICT VALIDATION**: Jika $\text{New Potential Debt} > \text{Reseller.CreditLimit}$, sistem **wajib menolak** transaksi dengan pesan error: `Credit limit exceeded. Limit: X, Outstanding: Y, New Order Debt: Z`.
4.  **Pemberian Izin Cetak**:
    *   Jika lolos limit, order langsung beralih status ke `IN_PRODUCTION` dengan status pembayaran `UNPAID` (jika bayar awal 0) atau `PARTIAL_PAID` (jika bayar DP sebagian), dan `amount_paid` disimpan sesuai nilai bayar awal.

---

## 8. Endpoint Fleksibel Update Status Order (V2)

*   **Method & URL**: `PATCH /api/v1/orders/:id/status`
*   **Request Body (JSON)**:
    ```json
    {
      "status": "DRAFT"
    }
    ```
*   **Matriks Transaksi Status yang Didukung**:
    *   `DRAFT` -> `PENDING_PAYMENT`, `CANCELLED`
    *   `PENDING_PAYMENT` -> `DRAFT`, `CANCELLED`
    *   `IN_PRODUCTION` -> `READY_FOR_PICKUP`, `CANCELLED`
    *   `READY_FOR_PICKUP` -> `COMPLETED` (hanya jika `payment_status == PAID`), `CANCELLED`
 
 
## 9. Endpoint Update Draft Order (V2)
 
*   **Method & URL**: `PUT /api/v1/orders/:id`
*   **Keterangan**: Digunakan untuk memperbarui data items, reseller, customer info, dan notes pada draft order. Hanya diperbolehkan jika status order saat ini masih `DRAFT`.
*   **Request Body (JSON)**: Sama dengan payload `POST /api/v1/orders/draft` (`CreateOrderReq`).
*   **Response (JSON)**: Sama dengan response `Save Job Entry Draft` dengan harga yang sudah dikalkulasi ulang.

