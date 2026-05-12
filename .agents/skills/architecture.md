# 🏗 Modular Clean Architecture

### 🧱 Principles
*   **Feature-Based Isolation**: Aplikasi dibagi per module/fitur di dalam folder `internal/` (e.g., `internal/user`, `internal/auth`).
*   **Self-Contained Layers**: Setiap module memiliki layer internal:
    *   `domain/`: Entity & Interface.
    *   `usecase/`: Business Logic.
    *   `repository/`: Database (GORM).
    *   `delivery/http/`: Handler & DTO.
*   **Scalability**: Memudahkan penambahan fitur baru tanpa mengganggu module lain.
*   **Refactor Friendly**: Menghapus fitur semudah menghapus satu folder module.
