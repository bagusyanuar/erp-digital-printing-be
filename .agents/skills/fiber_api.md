# 🚀 Fiber v3 & API Response

### 🛠 Technical Standards
*   **Fiber v3**: Memanfaatkan fitur `c.Bind().Body(&req)` untuk parsing request.
*   **Hybrid Token Strategy**:
    *   `access_token`: Dikirim di JSON body response.
    *   `refresh_token`: Dikirim via **HttpOnly Cookie** untuk keamanan dari XSS.
*   **Generic Response**: Wajib menggunakan explicit type hint `[any]` pada `response.Success` jika mengirim data `nil` untuk menghindari error type inference.

### 📦 Standard Response Format
*   Selalu gunakan helper di `pkg/response` (`Success`, `Created`, `Error`) untuk menjaga konsistensi format JSON API.
