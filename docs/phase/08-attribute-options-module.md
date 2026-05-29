# Phase 8: Attribute Options Module

## Goals
Implement Attribute Options (relational table `attribute_options`) to support pre-defined options for Attributes with type `options` (e.g., Sisi -> "1 Sisi", "2 Sisi").

## Tasks

### 1. Database Migration
- [x] Create `000007_create_attribute_options_table.up.sql` to create `attribute_options` table.
- [x] Create `000007_create_attribute_options_table.down.sql` to drop `attribute_options` table.

### 2. Domain Layer (Go)
- [x] Define `AttributeOption` model in `internal/product/domain/product.go`.
- [x] Update `Attribute` model to include `Options []AttributeOption` relation.

### 3. Repository & Usecase Layer
- [x] Update `Attribute` CRUD implementation in repository to support creating, updating, and deleting `AttributeOption` via GORM associations.
- [x] Ensure options are preloaded when fetching Attributes.

### 4. Delivery Layer (HTTP)
- [x] Update Attribute Request DTOs in `internal/product/delivery/http/dto/attribute_dto.go` to accept `options` field (list of strings/objects).
- [x] Update Attribute Response DTOs to include preloaded `options` list.
- [x] Verify endpoints with Bruno client.
