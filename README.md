# ERP Digital Printing - Backend

Backend system for ERP Digital Printing built with Go, Fiber v3, and GORM.

## 🚀 Tech Stack
- **Language**: Go (Golang)
- **Framework**: Fiber v3
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT (Access Token & HttpOnly Refresh Token)
- **RBAC**: Casbin
- **Containerization**: Docker & GHCR

## 🛠 Prerequisites
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL (if running locally)
- `migrate` tool (for database migrations)

## 📦 Quick Start (Docker - For Users)

If you just want to run the application without the source code, you can use the pre-built image from GitHub Container Registry:

1. Create a `.env` file from the example:
   ```bash
   cp .env.example .env
   ```
2. Run using Docker Compose:
   ```yaml
   # docker-compose.yml
   services:
     app:
       image: ghcr.io/bagusyanuar/erp-digital-printing-be:latest
       ports:
         - "8000:8000"
       env_file: .env
       depends_on:
         db:
           condition: service_healthy
     db:
       image: postgres:16-alpine
       environment:
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: password123
         POSTGRES_DB: db_erp_printing
       healthcheck:
         test: ["CMD-SHELL", "pg_isready -U postgres"]
         interval: 5s
         timeout: 5s
         retries: 5
   ```
3. Start the containers:
   ```bash
   docker compose up -d
   ```

## 👨‍💻 Local Development (For Contributors)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/bagusyanuar/erp-digital-printing-be.git
   cd erp-digital-printing-be
   ```

2. **Setup environment variables**:
   ```bash
   cp .env.example .env
   ```

3. **Database Migrations**:
   ```bash
   make migrate-up
   ```

4. **Run Application**:
   ```bash
   # Using Air for Hot Reload
   air
   
   # Or standard Go run
   go run cmd/api/main.go
   ```

## 🚀 Panduan Rilis (CI/CD Release Guide)

Repo ini memiliki alur rilis terpisah antara **Development** dan **Production** yang otomatis terintegrasi dengan GitHub Container Registry (GHCR) dan GitHub Actions.

### Ringkasan Alur

| Target | Format Tag Git | Workflow GitHub Actions | Hasil Docker Image (GHCR) | Deployment |
| :--- | :--- | :--- | :--- | :--- |
| **Development** | `dev-vX.Y.Z` | `.github/workflows/docker-dev.yml` | `dev-latest`, `dev-<sha>`, `dev-X.Y.Z` | Manual (CI Only) |
| **Production** | `vX.Y.Z` | `.github/workflows/docker-prod.yml` | `latest`, `X.Y.Z`, `major.minor` | Otomatis via SSH (CI/CD) |

---

### Cara Melakukan Release

Gunakan command `make` yang sudah disediakan:

#### 1. Mode Interaktif (Paling Direkomendasikan)
```bash
make release
```
Panduan memilih:
1. **Target:** `1) dev` atau `2) prod`
2. **Tipe Bump Versi:**
   - `0) initial` (contoh: `1.0.0` -> rilis versi base saat ini)
   - `1) patch` (contoh: `1.0.0` -> `1.0.1`)
   - `2) minor` (contoh: `1.0.0` -> `1.1.0`)
   - `3) major` (contoh: `1.0.0` -> `2.0.0`)
   - `4) custom` (masukkan versi manual)
3. Script akan otomatis:
   - Mengupdate versi di file `VERSION`.
   - Membuat git commit (`chore(release): ...`).
   - Membuat git tag beranotasi (`dev-v...` atau `v...`).
   - Menanyakan konfirmasi untuk push ke GitHub origin.

---

#### 2. Mode Shortcut (Cepat)

##### Rilis ke Development:
```bash
# Wizard dev
make release-dev

# Langsung bump
make release-dev type=patch
make release-dev type=minor
make release-dev type=2.2.0
```

##### Rilis ke Production:
```bash
# Wizard prod
make release-prod

# Langsung bump
make release-prod type=patch
make release-prod type=minor
make release-prod type=2.2.0
```

---

## 🛠️ GitHub Actions Secrets (Production Deployment)

Pastikan Secrets berikut telah diatur di **GitHub Repository > Settings > Secrets and variables > Actions**:

- `PROD_SERVER_HOST`: IP Address atau Hostname VPS Production.
- `PROD_SERVER_USER`: Username SSH server (contoh: `ubuntu`, `root`).
- `PROD_SSH_PRIVATE_KEY`: Private Key SSH server.
- `PROD_SERVER_PORT`: Port SSH (default: `22` jika kosong).
- `PROD_APP_PATH`: Folder docker compose di server (default: `/home/made-vps/made-deployment`).

---

## 🧹 Database Reset (Go-Live / Production Preparation)

Gunakan panduan ini untuk menghapus data testing/transaksi dummy sebelum go-live ke production tanpa menghapus Master Data (Users, RBAC, Produk, Kategori, Reseller, Supplier, Rekening Kas).

### SQL Script Reset Transaksi

Jalankan script berikut (melalui DBeaver / `psql`):

```sql
BEGIN;

-- 1. Kosongkan semua data transaksi (child -> parent cascade)
TRUNCATE TABLE 
    order_item_finishings,
    order_items,
    order_payments,
    orders,
    expense_items,
    expense_payments,
    expenses,
    fund_transfers,
    capital_transactions,
    cash_flows
CASCADE;

-- 2. Reset saldo akun kas/bank ke 0
UPDATE cash_accounts SET balance = 0;

COMMIT;
```

### Cara Eksekusi via DBeaver:
1. Buka SQL Editor: Klik kanan database/schema `public` ➔ **SQL Editor** ➔ **New SQL Script** (`Cmd + ]` / `Ctrl + ]`).
2. Paste script di atas.
3. Jalankan script: Tekan `Option + X` (Mac) atau `Alt + X` (Windows) / klik tombol **Execute SQL Script** di toolbar sebelah kiri.
4. Refresh tables: Tekan `F5`.

---

## 🖨️ Printer Agent (Print Thermal Lokal)

Printer Agent (`cmd/printer`) adalah aplikasi kecil yang dipasang di **PC kasir (Windows)**, bukan di VPS. Frontend yang dibuka di browser kasir mengirim perintah ESC/POS ke `http://localhost:9876`, lalu agent meneruskannya ke printer thermal melalui Windows Print Spooler.

```
Browser kasir (FE dari VPS) ──HTTP──▶ Printer Agent (localhost:9876) ──RAW──▶ Printer Thermal
```

Agent berjalan sebagai **Windows Service** (`ERPPrinterAgent`): otomatis jalan saat boot, tanpa jendela terminal, dan otomatis restart jika crash.

### 1. Build

Dari Mac/Linux/Windows di root repo:

```bash
GOOS=windows GOARCH=amd64 go build -o printer-agent.exe ./cmd/printer
```

### 2. Install Service

1. Copy `printer-agent.exe` ke folder permanen di PC kasir, contoh `C:\ERP\PrinterAgent\`.
2. Pastikan printer thermal sudah terpasang di Windows dengan nama **`PRINTER-POS`** (Settings ➔ Printers & scanners). Nama ini yang dipakai FE.
3. Buka **Command Prompt sebagai Administrator**, lalu:

```bat
cd C:\ERP\PrinterAgent
printer-agent.exe install
printer-agent.exe start
```

4. Cek agent berjalan: buka `http://localhost:9876/health` di browser, harus muncul `{"status":"ok", ...}`.

> ⚠️ Jangan pindahkan `.exe` setelah `install` (service menyimpan path-nya). Jika ingin pindah folder: `uninstall` ➔ pindahkan ➔ `install` lagi.

### 3. Perintah Service

Semua dijalankan di Command Prompt **Administrator**:

| Perintah | Fungsi |
|---|---|
| `printer-agent.exe install` | Daftarkan service ke Windows |
| `printer-agent.exe start` | Jalankan service |
| `printer-agent.exe stop` | Hentikan service |
| `printer-agent.exe restart` | Restart service |
| `printer-agent.exe uninstall` | Hapus service dari Windows |

Service juga bisa dikelola lewat `services.msc` ➔ **ERP Printer Agent**.

### 4. Uninstall Service

```bat
cd C:\ERP\PrinterAgent
printer-agent.exe stop
printer-agent.exe uninstall
```

Setelah itu folder `C:\ERP\PrinterAgent\` boleh dihapus.

### 5. Update ke Versi Baru

```bat
printer-agent.exe stop
:: timpa printer-agent.exe dengan hasil build terbaru
printer-agent.exe start
```

### 6. Cara Konek dari Frontend

Agent hanya listen di `127.0.0.1:9876`, jadi FE harus memanggil **`http://localhost:9876`** dari browser di PC yang sama.

| Method | Endpoint | Keterangan |
|---|---|---|
| `GET` | `/health` | Cek agent hidup |
| `POST` | `/print` | Kirim data ESC/POS ke printer |
| `GET` | `/print-test` | Mengembalikan contoh data ESC/POS (tidak langsung print) |

Payload `POST /print`:

```json
{
  "printer_name": "PRINTER-POS",
  "raw_data": "\u001b@Halo dunia\n\n\n\u001dV\u0000"
}
```

Contoh pemanggilan dari FE:

```ts
const response = await fetch("http://localhost:9876/print", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    printer_name: "PRINTER-POS",
    raw_data: rawData, // string ESC/POS
  }),
});

if (!response.ok) {
  const { error } = await response.json();
  throw new Error(error);
}
```

Mode debug (tanpa service): jalankan `printer-agent.exe` langsung (double-click / dari cmd), log akan tampil di terminal. Tutup dengan `Ctrl + C`.

### 7. Konfigurasi & Troubleshooting

- **Log**: tersimpan di `logs\printer_agent.log` di samping `.exe` (rotasi otomatis 10MB, simpan 5 file, 30 hari).
- **Ganti port**: set `PRINTER_PORT` di **System** environment variables (bukan User), lalu `printer-agent.exe restart`. Ingat sesuaikan URL di FE.
- **`Gagal membuka printer`**: nama printer tidak cocok atau tidak terlihat oleh akun service. Service berjalan sebagai `LocalSystem`, sehingga printer jaringan/shared yang di-add per user tidak terlihat. Solusi: pakai printer USB lokal yang terpasang untuk semua user, atau buka `services.msc` ➔ **ERP Printer Agent** ➔ tab **Log On** ➔ pilih akun user kasir ➔ restart service.
- **FE gagal konek (CORS / `Failed to fetch`)**: pastikan service status *Running* dan `http://localhost:9876/health` bisa dibuka di browser yang sama.

---

## 📖 Documentation
- [Docker Deployment Guide](docs/deployment/DOCKER.md)
- [Database Schema](docs/databases/)
- [API Collection (Bruno)](docs/api/bruno/)

## 📜 License
This project is licensed under the MIT License.
