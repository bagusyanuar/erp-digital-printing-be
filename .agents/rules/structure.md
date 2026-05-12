# 📂 Modular Clean Architecture - ERP Digital Printing

### 🏗 Modular Hierarchy
Aplikasi dibagi per **Module** di dalam folder `internal/`. Setiap module mengimplementasikan Clean Architecture sendiri.

*   **`internal/<module>/domain/`**: Business entities & interfaces.
*   **`internal/<module>/usecase/`**: Business logic.
*   **`internal/<module>/repository/`**: Database persistence (GORM).
*   **`internal/<module>/delivery/http/`**: Fiber handlers & routing.
*   **`internal/shared/`**: Global components (Config, DB, Logger, Container).

### 💉 Dependency Injection (DI)
*   **Location**: Wiring antar layer per module dilakukan di Constructor masing-masing layer.
*   **Container**: Wiring antar module dilakukan di `internal/shared/container/`.
    *   Wajib gunakan file terpisah per module (misal: `user_container.go`).
    *   `container.go` bertindak sebagai **Aggregator**.
*   **Package Alias**: Wajib gunakan alias saat import handler di `container.go` (contoh: `userHttp`) untuk menghindari konflik penamaan package `http`.
*   **Pattern**: Gunakan Constructor (`New...`) di tiap layer.

### 🚀 Bootstrap Flow
1.  Initialize **Config** & **Logger**.
2.  Open **DB Connection**.
3.  Call **Container** untuk wiring semua module.
4.  Setup **Fiber App** & **Middleware**.
5.  Register **Routes** dari Module Handler.
6.  Graceful **Shutdown**.
