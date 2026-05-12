# 💉 Dependency Injection: Aggregator Pattern

### 🏗 Patterns
*   **Modular Container**: Wiring DI dilakukan per file module (e.g., `user_container.go`) di dalam package `container`.
*   **Aggregator Container**: `container.go` bertindak sebagai entry point tunggal yang menggabungkan semua module container.
*   **Dual JWT Instances**: Menggunakan instance `JWTUtil` terpisah untuk Access Token dan Refresh Token jika menggunakan secret/durasi berbeda.
*   **Constructor Injection**: Wajib menggunakan fungsi `New...` untuk instansiasi setiap layer.

### 🔌 Aliasing
*   **Package Alias**: Wajib menggunakan alias (e.g., `userHttp`, `authHttp`) saat import handler di container untuk menghindari naming conflict pada package `http`.
