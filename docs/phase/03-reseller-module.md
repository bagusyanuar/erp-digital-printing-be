# Phase 3: Reseller Module Implementation

## Goals
Implement Master Data for Resellers (Biro/Partners) to support B2B transactions in the ERP system.

## Tasks

### 1. Research & Brainstorming
- [x] Brainstorm reseller table schema (id, name, email, phone, address, credit_limit).
- [x] Decide on UUID for Primary Key.
- [x] Create DBML schema in `docs/databases/reseller.schema.dbml`.

### 2. Database Implementation
- [x] Create SQL migrations (Up/Down) in `migrations/`.
- [x] Run migrations using `make migrate-up`.
- [x] Verify table structure in PostgreSQL.

### 3. Domain Layer
- [x] Define `Reseller` entity with GORM tags and JSON tags.
- [x] Define `ResellerRepository` and `ResellerUsecase` interfaces.
- [x] Implement `BeforeCreate` hook for UUID generation.

### 4. Repository & Usecase Layer
- [x] Implement GORM Repository with soft delete support.
- [x] Implement filtering logic (Search name, email, phone, address).
- [x] Implement pagination and sorting logic using shared `request.PaginationParam`.
- [x] Implement Usecase logic for CRUD.

### 5. Delivery Layer (HTTP)
- [x] Define Request/Response DTOs for Reseller.
- [x] Implement Fiber v3 Handlers for CRUD operations.
- [x] Implement Query Parameter binding for pagination and search.
- [x] Standardize API responses using `pkg/response`.

### 6. Integration & Security
- [x] Register module in DI Container (`reseller_container.go`).
- [x] Register protected routes in `bootstrap/app.go`.
- [x] Apply Auth and RBAC middleware to Reseller endpoints.

### 7. Documentation
- [x] Create Bruno API Collection for Reseller module.
- [x] Add detailed documentation for filtering and pagination parameters.

## Technical Details
- **Architecture**: Clean Architecture (Modular).
- **Primary Key**: UUID v4.
- **Filtering**: `ILIKE` (Case-Insensitive Search).
- **Pagination**: Centalized `request.PaginationParam`.
- **Authorization**: Protected by JWT & Casbin RBAC.
