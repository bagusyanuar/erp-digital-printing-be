# 🏗 Clean Architecture Rules - ERP Digital Printing

### 🧱 Layering Standards
*   **Domain**: `internal/domain`. Isi: Entity (struct), Interface (Repository/Usecase), & Custom Errors. NO dependencies to other layers.
*   **Usecase**: `internal/usecase`. Isi: Business logic. Tergantung hanya pada Domain. Inject Repository via Interface.
*   **Repository**: `internal/repository`. Isi: GORM implementation. Tergantung pada Domain & GORM.
*   **Delivery**: `internal/delivery/http`. Isi: Fiber Handlers & Middleware. Tergantung pada Usecase.

### 🚦 Dependency Flow
*   Flow: Handler -> Usecase -> Repository.
*   Dependency Injection (DI): Wajib pakai Constructor (`New...`) untuk semua layer.
*   Dilarang keras: Repository panggil Usecase, atau Usecase panggil Handler.

### 📦 Data Transfer Objects (DTO)
*   **Request**: Pakai struct untuk validasi input Fiber (misal: `CreateProductReq`).
*   **Response**: Pakai struct untuk output JSON seragam.
*   Domain Entity dilarang bocor langsung ke client jika ada sensitive data (password, dll).

### 🛠 Technical Rules (Go & Fiber)
*   **Framework**: Fiber v2.
*   **Database**: GORM v2.
*   **Logging**: Zap Logger (inject ke usecase/handler).
*   **Error Handling**: Gunakan `fiber.Error` atau custom domain error yang di-map di handler.
*   **Context**: Selalu pass `context.Context` dari Fiber ke Usecase sampai Repository.
