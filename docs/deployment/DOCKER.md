# Deployment Guide - Docker & GHCR

Panduan untuk membangun Docker image secara mandiri dan menjalankan aplikasi dalam container.

## 1. Membangun Image Secara Lokal

Jika Anda ingin mencoba membangun (build) image secara manual di mesin lokal:

```bash
docker build -t erp-app:local .
```

---

## 2. Menjalankan Container Mandiri (Standalone)

Jika Anda ingin menjalankan backend di dalam Docker tetapi menggunakan **Database Lokal (DBngin/Native)** di Mac:

### Langkah 1: Persiapan .env
Pastikan `.env` Anda menggunakan host khusus agar container bisa mengakses localhost Mac:
```env
DB_HOST=host.docker.internal
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=db_erp_printing
```

### Langkah 2: Jalankan Container
```bash
docker run -d \
  --name erp-backend \
  -p 8000:8000 \
  --env-file .env \
  --add-host=host.docker.internal:host-gateway \
  erp-app:local
```

---

## 3. CI/CD - GitHub Container Registry (GHCR)

Aplikasi otomatis dibangun dan dikirim ke GHCR ketika ada **Git Tag** baru.

### Cara Melakukan Release
1. Push kode ke branch `main`.
2. Buat tag baru:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. Image akan tersedia di: `ghcr.io/bagusyanuar/erp-digital-printing-be:latest`

---

## 4. Orchestration (Production)

Untuk keperluan deployment skala produksi (BE + FE + Nginx + DB), silakan merujuk ke repository deployment terpisah:
👉 `erp-digital-printing-deploy`

---

## 5. Remote Migration & Seeding (VPS)

Jika aplikasi sudah berjalan di VPS dan ingin melakukan migrasi dari Mac:

### Step 1: SSH Tunnel
```bash
ssh -L 5433:localhost:5432 user@ip-vps
```

### Step 2: Jalankan Command
```bash
make migrate-up DB_PORT=5433
make db-seed DB_PORT=5433
```
