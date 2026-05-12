# 💡 Go Backend Engineering Skills - ERP Digital Printing

### 🏗 Architecture: Modular Clean Architecture
*   **Pattern**: Membagi aplikasi berdasarkan fitur/module (e.g., `internal/user`).
*   **Layering**: Setiap module memiliki layer `domain`, `usecase`, `repository`, dan `delivery/http`.
*   **Isolation**: Setiap module bersifat self-contained untuk memudahkan skalabilitas dan maintenance.

### 💉 Dependency Injection: Aggregator Pattern
*   **Modular Wiring**: Wiring DI dipecah per file module (e.g., `user_container.go`) di dalam package `container`.
*   **Aggregator**: `container.go` hanya bertindak sebagai pengumpul (aggregator) dari semua module container.
*   **Aliasing**: Menggunakan alias package (e.g., `userHttp`) saat import handler di container untuk menghindari naming conflict package `http`.

### 🚀 Fiber v3 & API Response
*   **v3 Features**: Pemanfaatan `c.Bind()` untuk request handling yang lebih modern.
*   **Generic Response**: Wajib menggunakan explicit type hint `[any]` pada `response.Success` jika mengirim data `nil` untuk menghindari error type inference.
*   **Consistency**: Selalu gunakan `pkg/response` untuk format JSON yang seragam.

### 🗄 Database & Persistence
*   **UUID v4**: Standardisasi primary key menggunakan UUID v4 (`gen_random_uuid()`).
*   **GORM Best Practices**: 
    *   Wajib pass `context.Context` lewat `.WithContext(ctx)`.
    *   Penggunaan Soft Delete (`deleted_at`) secara konsisten.
*   **Migrations**: Manajemen skema menggunakan `golang-migrate` (SQL-based).

### 🛠 Coding Style & Standards
*   **Golang 1.26+**: Memanfaatkan fitur terbaru Go.
*   **Clean Code**: Constructor pattern (`New...`) untuk semua layer.
*   **No Interface{}**: Selalu gunakan `any` sesuai rule project.
