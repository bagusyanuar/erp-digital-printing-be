# 📂 Project Structure & DI Rules - ERP Digital Printing

### 🏗 Folder Hierarchy
*   **`cmd/`**: Entry point aplikasi (misal: `cmd/api/main.go`).
*   **`internal/domain/`**: Business entities & interfaces (Pure Go, no external libs).
*   **`internal/usecase/`**: Business logic.
*   **`internal/repository/`**: Database persistence (GORM).
*   **`internal/delivery/http/`**: Fiber handlers & routing.
*   **`internal/shared/`**: Global components (Config, DB, Logger).
*   **`internal/shared/container/`**: Dependency Injection container (wiring).

### 💉 Dependency Injection (DI)
*   **Location**: Semua wiring logic wajib ada di `internal/shared/container`.
*   **Pattern**: Gunakan Constructor (`New...`) di tiap layer.
*   **Registration**: Module baru wajib didaftarkan di container agar bisa di-bootstrap di `main.go`.
*   **Anti-Pattern**: Dilarang melakukan hard-code instansiasi database atau repository di dalam Usecase. Semua wajib di-inject via interface.

### 🚀 Bootstrap Flow
1.  Initialize **Config** & **Logger**.
2.  Open **DB Connection**.
3.  Call **Container** untuk wiring semua module.
4.  Setup **Fiber App** & **Middleware**.
5.  Register **Routes** dari Handler.
6.  Graceful **Shutdown**.
