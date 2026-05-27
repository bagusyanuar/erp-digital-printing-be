# Phase 4: Category Module Implementation

## Goals
Implement Master Data for Categories to classify products, materials, and other classifications in the ERP system.

## Tasks

### 1. Research & Brainstorming
- [x] Brainstorm category table schema (id, name).
- [x] Decide on UUID for Primary Key.
- [x] Create DBML schema in `docs/databases/category.schema.dbml`.

### 2. Database Implementation
- [x] Create SQL migrations (Up/Down) in `migrations/`.
- [x] Run migrations using `make migrate-up`.
- [x] Verify table structure in PostgreSQL.

### 3. Domain Layer
- [x] Define `Category` entity with GORM tags and JSON tags.
- [x] Define `CategoryRepository` and `CategoryUsecase` interfaces.
- [x] Implement `BeforeCreate` hook for UUID generation.

### 4. Repository & Usecase Layer
- [x] Implement GORM Repository with soft delete support.
- [x] Implement filtering logic (Search name).
- [x] Implement pagination and sorting logic using shared `request.PaginationParam`.
- [x] Implement Usecase logic for CRUD.

### 5. Delivery Layer (HTTP)
- [x] Define Request/Response DTOs for Category.
- [x] Implement Fiber v3 Handlers for CRUD operations.
- [x] Implement Query Parameter binding for pagination and search.
- [x] Standardize API responses using `pkg/response`.

### 6. Integration & Security
- [x] Register module in DI Container.
- [x] Register protected routes in HTTP bootstrap.
- [x] Apply Auth middleware to Category endpoints.

### 7. Documentation
- [x] Create Bruno API Collection for Category module.

## Technical Details
- **Architecture**: Clean Architecture (Modular).
- **Primary Key**: UUID v4.
- **Filtering**: `ILIKE` (Case-Insensitive Search on name).
- **Pagination**: Centralized `request.PaginationParam`.
- **Authorization**: Protected by JWT.
