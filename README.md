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

## 📖 Documentation
- [Docker Deployment Guide](docs/deployment/DOCKER.md)
- [Database Schema](docs/databases/)
- [API Collection (Bruno)](docs/api/bruno/)

## 📜 License
This project is licensed under the MIT License.
