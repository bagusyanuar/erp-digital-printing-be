# Product Module Planning — ERP Digital Printing

Dokumen ini mendefinisikan arsitektur, skema database, sistem pricing, dan aturan bisnis untuk modul produk yang menggunakan kombinasi **EAV (Entity-Attribute-Value)**, **Tiered Pricing (Harga Berjenjang & Level Customer)**, serta **Product Bundling**.

---

## 1. Arsitektur Database & EAV

Penerapan EAV (Entity-Attribute-Value) dihubungkan pada tingkat **Variant**, bukan langsung pada Product. Hal ini memberikan fleksibilitas tinggi untuk produk digital printing yang memiliki spesifikasi kustom berbeda untuk tiap variannya (misal: bahan, finishing, ukuran).

### A. Tabel `products` (Entity Utama)
Menyimpan data dasar produk yang bersifat global.
- `id` (UUID, PK)
- `category_id` (UUID, FK -> `categories`)
- `name` (VARCHAR)
- `sku` (VARCHAR, UNIQUE)
- `uom` (VARCHAR) — Unit of Measurement (pcs, meter, rim, dll)
- `base_price` (DECIMAL) — Harga dasar acuan awal
- `created_at`, `updated_at`, `deleted_at` (Timestamps)

### B. Tabel `attributes` (Attribute Definition)
Menyimpan definisi atribut spesifikasi produk.
- `id` (UUID, PK)
- `name` (VARCHAR) — Nama atribut (misal: Bahan, Finishing, Laminasi)
- `code` (VARCHAR, UNIQUE) — Kode unik (misal: bahan, finishing, laminasi)
- `value_type` (VARCHAR) — Tipe data nilai atribut (`text`, `number`, `boolean`, `options`)
- `created_at`, `updated_at`, `deleted_at` (Timestamps)

### C. Tabel `product_variants` (Variant Entity)
Menyimpan variasi spesifik dari sebuah produk. Setiap produk minimal memiliki 1 default variant jika produk tersebut tidak memiliki varian kustom.
- `id` (UUID, PK)
- `product_id` (UUID, FK -> `products`)
- `variant_name` (VARCHAR) — Nama varian (misal: "A3+ Art Paper 260gr", "Default")
- `additional_cost` (DECIMAL, DEFAULT 0) — Biaya tambahan untuk varian ini
- `is_default` (BOOLEAN, DEFAULT false) — Flag penanda varian default
- `created_at`, `updated_at`, `deleted_at` (Timestamps)

### D. Tabel `product_attribute_values` (EAV Value Link)
Menghubungkan Variant dengan Attribute beserta nilainya.
- `id` (UUID, PK)
- `product_variant_id` (UUID, FK -> `product_variants`)
- `attribute_id` (UUID, FK -> `attributes`)
- `value` (TEXT) — Nilai spesifik dari atribut (misal: "Art Paper 260gr", "Glossy")
- `created_at`, `updated_at`, `deleted_at` (Timestamps)

---

## 2. Tingkat & Skema Harga (Pricing Matrix)

Sistem ERP mendukung harga berjenjang (volume discount) yang dibedakan berdasarkan level customer (misal: End User, Reseller).

### A. Tabel `customer_levels`
Menyimpan tingkatan level customer. Master data ini juga dihubungkan dengan tabel `resellers` (`resellers.customer_level_id -> customer_levels.id`) untuk penentuan otomatis skema harga saat transaksi.
- `id` (UUID, PK)
- `name` (VARCHAR) — Nama level (misal: "End User", "Reseller")
- `discount_percentage` (DECIMAL, DEFAULT 0) — Diskon global (jika diperlukan)
- `created_at`, `updated_at`, `deleted_at` (Timestamps)


### B. Tabel `price_tiers` (Tiered Pricing Matrix)
Menyimpan harga per unit berdasarkan jumlah kuantitas order (`min_qty` s/d `max_qty`) dan level customer.
- `id` (UUID, PK)
- `product_variant_id` (UUID, FK -> `product_variants`)
- `customer_level_id` (UUID, FK -> `customer_levels`)
- `min_qty` (INT) — Batas minimum kuantitas
- `max_qty` (INT, NULLABLE) — Batas maksimum kuantitas (NULL berarti tak terhingga/lebih dari min_qty)
- `price_per_unit` (DECIMAL) — Harga per unit pada tier ini
- `created_at`, `updated_at`, `deleted_at` (Timestamps)
- *Constraint*: `UNIQUE (product_variant_id, customer_level_id, min_qty)` untuk mencegah tumpang tindih tier dasar.

### C. Rumus Perhitungan Harga
$$\text{Final Price} = \text{price\_tiers.price\_per\_unit} + \text{product\_variants.additional\_cost}$$

---

## 3. Sistem Bundling (Paket Produk)

Mendukung penjualan paket gabungan dari beberapa varian produk dengan harga paket tetap (fixed price) yang ditentukan oleh admin.

### A. Tabel `bundles`
Menyimpan informasi paket/bundling.
- `id` (UUID, PK)
- `name` (VARCHAR) — Nama paket (misal: "Paket Promosi UMKM")
- `sku` (VARCHAR, UNIQUE)
- `base_price` (DECIMAL) — Harga jual paket tetap yang di-input admin
- `created_at`, `updated_at`, `deleted_at` (Timestamps)

### B. Tabel `bundle_items`
Menyimpan item produk/varian penyusun paket tersebut.
- `id` (UUID, PK)
- `bundle_id` (UUID, FK -> `bundles`)
- `product_variant_id` (UUID, FK -> `product_variants`)
- `qty` (INT) — Jumlah kuantitas varian produk di dalam paket
- `created_at`, `updated_at`, `deleted_at` (Timestamps)

---

## 4. Aturan Bisnis & Validasi Utama

1. **Auto Default Variant**:
   Saat produk baru dibuat, sistem wajib membuat 1 record di `product_variants` dengan flag `is_default = true` dan `additional_cost = 0`. Ini memastikan `price_tiers` bisa langsung diarahkan ke varian default jika produk tidak memiliki variasi kompleks.
2. **Additional Cost Tetap**:
   `additional_cost` pada varian bersifat statis dan langsung ditambahkan ke harga per unit yang didapat dari tabel `price_tiers`. Nilai default adalah `0`.
3. **Fixed Bundle Price**:
   Harga jual paket (`bundles.base_price`) di-input langsung oleh admin dan bersifat absolut. Sistem tidak menghitung diskon otomatis dari total harga satuan item di dalamnya, melainkan langsung menggunakan harga paket tersebut saat checkout.
4. **Validasi Range Qty**:
   Saat input `price_tiers`, sistem harus memastikan tidak ada overlapping range kuantitas untuk `product_variant_id` dan `customer_level_id` yang sama.
