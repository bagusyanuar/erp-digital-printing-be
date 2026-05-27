# Phase 7: Attribute Master Data Module

## Goals
Implement Master Data for Attributes (EAV metadata definition) to standardized specifications (e.g. Bahan, Ukuran, Finishing) across all product variants.

## Tasks

### 1. Domain Layer (Go)
- [x] Define Repository & Usecase Interfaces for Attributes in `internal/product/domain/product.go` (already partially defined as struct).
- [x] Implement `Attribute` CRUD models.

### 2. Repository & Usecase Layer
- [x] Implement GORM Repository for `Attribute` with soft delete.
- [x] Implement Usecase logic for managing master attributes.
- [x] Ensure attribute code is unique and automatically slugified (e.g., "Laminasi Doff" -> "laminasi_doff") during create/update.

### 3. Delivery Layer (HTTP)
- [x] Define Request/Response DTOs for Attribute.
- [x] Implement Fiber v3 Handlers for CRUD:
  * `POST /attributes` (Create attribute)
  * `GET /attributes` (List all with pagination/search)
  * `GET /attributes/:id` (Get by ID)
  * `PUT /attributes/:id` (Update attribute)
  * `DELETE /attributes/:id` (Soft delete attribute)
- [x] Register Handlers & dependency injection in DI Container & Router bootstrap.
- [x] Create Bruno API Collection for Attribute module.

## Technical Details
- **Code Generation**: Automated conversion of name to lowercase unique code (e.g. Name: "Laminasi Kartu" -> Code: "laminasi_kartu").
- **Value Types**: Restricted validation on `value_type` using `oneof=text number boolean options`.
