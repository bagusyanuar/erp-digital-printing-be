# 🛠 Coding Style & Standards

### 📝 Go Standards
*   **Version**: Go 1.26+.
*   **Syntax**: Gunakan `any` sebagai pengganti `interface{}` (sesuai rule ERP Digital Printing).
*   **Constructors**: Setiap struct di layer manapun wajib memiliki constructor function `New...`.

### 🚦 Error Handling
*   **Domain Errors**: Definisikan custom error di level domain jika perlu.
*   **Fiber Errors**: Gunakan `response.Error` di level delivery untuk standarisasi output error ke client.
