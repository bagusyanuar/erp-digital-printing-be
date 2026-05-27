# Phase 6: Reseller Integration & Pricing Adjustment

## Goals
Integrate the Reseller module with the new `customer_levels` schema and adjust CRUD reseller operations to utilize tiered pricing levels.

## Tasks

### 1. Database Alter Migration
- [x] Create SQL migration for altering `resellers` table:
  * Add `customer_level_id` column as UUID.
  * Define Foreign Key constraint to `customer_levels(id)` with `ON DELETE RESTRICT`.
- [x] Execute migration manually (`make migrate-up`).

### 2. Domain & CRUD Adjustment (Go)
- [x] Update `Reseller` domain model in `internal/reseller/domain/reseller.go`:
  * Add `CustomerLevelID` field.
  * Add `CustomerLevel` relation (preload support if needed).
- [x] Adjust `ResellerDTO` in `internal/reseller/delivery/http/dto/reseller_dto.go`:
  * Add `customer_level_id` validation in Create/Update Request DTOs.
  * Map `customer_level_id` in response DTO.
- [x] Update Usecase & Repository Layer:
  * Automate assigning default `"Reseller"` customer level ID upon creating a new reseller if none provided.
- [x] Verify HTTP handlers and adjust Bruno integration tests.

## Technical Details
- **Migration Alter Strategy**: 
  1. Add column `customer_level_id` as nullable first, OR set default value to the existing `"Reseller"` ID.
  2. Perform data backfill (update existing resellers to use `"Reseller"` level ID).
  3. Set column `customer_level_id` to `NOT NULL` and add FK constraint.
- **Go Helper**: Use a dedicated shared function to retrieve default `"Reseller"` and `"End User"` level UUIDs easily.
