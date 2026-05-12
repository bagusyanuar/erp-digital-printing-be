# 🗄 Database & GORM Rules - ERP Digital Printing

### 🧬 Model Definition
*   **Naming**: Nama tabel snake_case, plural (misal: `products`, `orders`).
*   **Primary Key**: Selalu pakai `ID uuid.UUID` atau `uint` auto-increment sesuai kebutuhan.
*   **Timestamps**: Wajib ada `created_at` dan `updated_at`.

### 🗑 Soft Deletes (`deleted_at`)
*   **Selective Implementation**: Hanya tambahkan `gorm.DeletedAt` pada tabel master data atau transaksi krusial (misal: `products`, `customers`, `orders`).
*   **Exclusion**: Tabel log, session, atau cache tidak perlu soft delete.
*   **Usage**: Gunakan field `DeletedAt gorm.DeletedAt `index` ` di struct model.

### 🏷 GORM Tags
*   Wajib definisikan constraint di tag: `gorm:"primaryKey"`, `gorm:"not null"`, `gorm:"uniqueIndex"`.
*   Relasi: Gunakan `foreignKey` dan `references` secara eksplisit agar tidak bingung.

### 🚀 Migrations
*   Gunakan `db.AutoMigrate()` hanya di fase development awal.
*   Untuk production, disarankan pakai tool migration terpisah (misal: `golang-migrate`) untuk tracking history.
