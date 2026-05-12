# 🌐 API Documentation (Bruno)

### 📂 Standards
*   **Git-Based**: Dokumentasi API harus sinkron dengan repo, disimpan di `docs/api/bruno/`.
*   **Module-Based**: Request dikelompokkan dalam folder sesuai module-nya.
*   **Sample-Rich**: Bagian `docs` di file `.bru` wajib berisi deskripsi endpoint dan sample response (Success & Error).

### ⚙️ Automation & Environments
*   **Variables**: Gunakan `{{base_url}}` untuk fleksibilitas ganti environment.
*   **Environments**: Definisi environment (Local, Staging, Prod) disimpan di sub-folder `environments/`.
