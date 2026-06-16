# Phase 15: Supplier Module

## Goals
Implement the `suppliers` table to manage vendor master data, provide CRUD API endpoints, and prepare fields for integration with purchase entries, expenses, and accounts payable (hutang).

## Tasks

### 1. Database Schema & Migration
- [x] Create migration files for the new table: `migrations/xxxxxx_create_suppliers_table.up.sql` and `migrations/xxxxxx_create_suppliers_table.down.sql`.
- [x] Table structure matches `docs/databases/supplier.schema.dbml` for `suppliers`.

### 2. Domain Model & Repository
- [x] Create `internal/supplier/domain/supplier.go` to model `Supplier` struct.
- [x] Implement query/mutator interfaces for repository.
- [x] Implement SQL queries with transaction support for inserting, updating, soft-deleting, and listing suppliers.

### 3. Usecase
- [x] Implement logic to manage suppliers:
  - Validate that name is present and not empty.
  - Handle search by name/contact/email and pagination for listing.
  - Check if supplier name is unique.

### 4. Handler & Routes
- [x] **Suppliers Routes:**
  - [x] `POST /api/v1/suppliers`
  - [x] `GET /api/v1/suppliers` (with search & pagination support)
  - [x] `GET /api/v1/suppliers/:id`
  - [x] `PUT /api/v1/suppliers/:id`
  - [x] `DELETE /api/v1/suppliers/:id`

## Technical Details
- Use `any` instead of `interface{}`.
- Logging using Zap logger.
- Integrate with PostgreSQL via GORM.
