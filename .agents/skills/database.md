# 🗄 Database & Persistence (GORM)

### 🆔 Primary Keys
*   **UUID v4**: Standardisasi PK menggunakan UUID v4.
*   **Application Level Generation**: Menggunakan GORM hook `BeforeCreate` dengan `uuid.New()` untuk generate ID di level Go sebelum masuk ke DB.

### 🛠 GORM Best Practices
*   **Context Awareness**: Wajib mempassing `context.Context` ke setiap query menggunakan `.WithContext(ctx)`.
*   **Soft Delete**: Menggunakan field `deleted_at` (type `*time.Time` dengan index GORM) untuk semua table master.
*   **DSN Format**: Gunakan format URL (`postgres://user:pass@host:port/db`) agar lebih robust menangani password kosong atau karakter spesial.

### 🚀 Migrations
*   **Golang-Migrate**: Mengelola skema database secara eksplisit menggunakan file `.sql` (`up` dan `down`).
