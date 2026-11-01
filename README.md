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
- `PROD_APP_PATH`: Folder docker compose di server (default: `/var/www/prod-app`).

---

## 📖 Documentation
- [Docker Deployment Guide](docs/deployment/DOCKER.md)
- [Database Schema](docs/databases/)
- [API Collection (Bruno)](docs/api/bruno/)

## 📜 License
This project is licensed under the MIT License.
