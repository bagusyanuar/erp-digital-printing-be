# Phase 2: RBAC Implementation

## Goals
Implement a robust Role-Based Access Control (RBAC) system supporting multiple roles per user using Casbin and GORM.

## Tasks

### 1. Research & Brainstorming
- [x] Brainstorm RBAC schema with multi-role support.
- [x] Decide on Casbin as the authorization engine.
- [x] Define DBML schema for RBAC tables.

### 2. Setup & Configuration
- [x] Install `casbin/v2` and `gorm-adapter/v3`.
- [x] Create `configs/rbac_model.conf` for Casbin RBAC model.
- [x] Update environment variables (`CASBIN_MODEL_PATH`).
- [x] Setup Casbin Helper in `pkg/casbin`.

### 3. Database Implementation
- [x] Create GORM models (domain).
- [x] Create SQL migrations in `migrations/` using `golang-migrate`.
- [x] Run migrations.
- [x] Seed initial roles and permissions.

### 4. Authorization Logic
- [x] Create RBAC Middleware for Fiber.
- [x] Map HTTP methods to granular actions (create, read, update, delete).
- [x] Integrate role extraction from JWT claims.

### 5. Integration
- [x] Protect existing User routes with RBAC.
- [x] Implement role assignment logic in User module.
- [ ] Validate multi-role enforcement.

### 6. RBAC Management API
- [x] Implement Repository for Role & Permission management.
- [x] Implement Usecase with Casbin auto-sync logic.
- [x] Implement HTTP Handlers (CRUD Roles, CRUD Permissions, Assign Role to User).
- [x] Protect RBAC endpoints (only for `administrator`).

### 7. Finalization
- [x] Integrate into DI Container.
- [ ] Validate wildcard access for `administrator`.
- [ ] Validate granular access for `admin` and `designer`.

## Technical Details
- **Engine**: Casbin (RBAC Model).
- **Storage**: PostgreSQL (via Gorm Adapter).
- **Schema**: Modular tables synced to `casbin_rule`.
