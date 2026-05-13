# Deployment Guide - Docker & GHCR

Panduan untuk menjalankan aplikasi ERP Digital Printing menggunakan Docker dan otomatisasi deployment ke GitHub Container Registry (GHCR).

## 1. Local Development (HTTPS)

Aplikasi dijalankan menggunakan Docker Compose dengan Nginx sebagai Reverse Proxy untuk mendukung HTTPS lokal.

### Persiapan SSL (mkcert)
Gunakan `mkcert` untuk membuat sertifikat SSL yang valid di lokal:

```bash
# Buat folder cert jika belum ada
mkdir -p docker/nginx/cert

# Generate sertifikat wildcard
mkcert -cert-file docker/nginx/cert/made-printing.local.pem \
       -key-file docker/nginx/cert/made-printing.local-key.pem \
       "*.made-printing.local" made-printing.local localhost 127.0.0.1
```

### Domain Mapping
Tambahkan domain berikut ke file `/etc/hosts` Anda:

```text
127.0.0.1 api.made-printing.local
```

### Menjalankan Aplikasi
```bash
docker compose up --build
```
Aplikasi dapat diakses di: `https://api.made-printing.local`

---

## 2. CI/CD - GitHub Container Registry (GHCR)

Aplikasi dikonfigurasi untuk otomatis membangun dan mengirim (push) Docker image ke GHCR hanya ketika ada **Git Tag** baru.

### Cara Melakukan Release
1. Pastikan kode sudah di-push ke branch `main`.
2. Buat tag baru dengan format `v*.*.*`:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. GitHub Actions akan otomatis mendeteksi tag tersebut dan memulai proses build.
4. Image hasil build akan tersedia di: `ghcr.io/<username>/erp-digital-printing-be`

### Tagging Strategy
Setiap build dari Git Tag akan menghasilkan image dengan tag:
- `v1.0.0` (sesuai versi tag)
- `1.0` (minor version)
- `latest` (selalu menunjuk ke versi tag terbaru)

---

## 3. Perintah Dasar Docker

- **Melihat Log**:
  ```bash
  docker compose logs -f app
  ```
- **Menghentikan Aplikasi**:
  ```bash
  docker compose down
  ```
- **Membersihkan Image & Volume**:
  ```bash
  docker compose down -v --rmi all
  ```
