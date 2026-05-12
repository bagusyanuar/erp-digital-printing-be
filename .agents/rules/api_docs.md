# 📖 API Documentation Rules (Bruno) - ERP Digital Printing

### 📂 Structure & Location
*   **Path**: Semua file `.bru` wajib berada di `docs/api/bruno/`.
*   **Grouping**: Kelompokkan request berdasarkan module (e.g., `Auth/`, `User/`, `Product/`).
*   **Naming**: Nama file harus deskriptif (e.g., `Login.bru`, `Create User.bru`).

### 📝 Documentation Standards (Inside .bru)
Setiap request wajib menyertakan bagian `docs` yang berisi:
1.  **Description**: Penjelasan singkat fungsi endpoint.
2.  **Auth/Behavior**: Penjelasan jika ada behavior khusus (misal: "Refresh token dikirim via cookie").
3.  **Sample Response**: 
    *   Wajib ada **Sample Success Response** (JSON).
    *   Wajib ada **Sample Error Response** (JSON) untuk case umum (400, 401, 404).

### 🌐 Environment & Variables
*   **No Hardcoding**: Dilarang menulis `http://localhost:8000` langsung di URL request.
*   **Base URL**: Selalu gunakan variabel `{{base_url}}` dari environment Bruno.
*   **Environment File**: Semua environment variable disimpan di `docs/api/bruno/environments/`.

### 🚀 Update Policy
*   Setiap ada perubahan logic/field di API (Request atau Response), file `.bru` terkait wajib langsung di-update agar sinkron dengan code backend.
