# Phase 5: Product Module Implementation (Core & EAV)

## Goals
Implement core product schema with EAV (Entity-Attribute-Value), Tiered Pricing (Price Tiers), and Bundling system.

## Tasks

### 1. Database & Migrations
- [x] Create database migration for product module:
  * `customer_levels` (Master data)
  * `products` (Master data)
  * `attributes` (EAV Attributes)
  * `product_variants` (Product variations)
  * `product_attribute_values` (EAV Values)
  * `price_tiers` (Tiered pricing)
  * `bundles` (Fixed-price bundles)
  * `bundle_items` (Bundle components)
- [x] Create database seeder for default `customer_levels` (`"End User"` & `"Reseller"`).
- [x] Execute migration manually (`make migrate-up`).

### 2. Domain & Core Entities (Go)
- [x] Define Go domain structs with GORM tags:
  * `CustomerLevel`
  * `Product`
  * `Attribute`
  * `ProductVariant`
  * `ProductAttributeValue`
  * `PriceTier`
  * `Bundle`
  * `BundleItem`
- [x] Implement `BeforeCreate` GORM hook for UUID generation in all entities.
- [x] Define Repository & Usecase Interfaces for Products & Bundles.

### 3. Core Product CRUD & EAV Engine
- [ ] Implement Product Repository (GORM) with preloading of variants and EAV attribute values.
- [ ] Implement Usecase logic for Product management (creating default variant upon product creation).
- [ ] Create HTTP Handlers and DTOs for Products & Variants.
- [ ] Register module in DI Container & Router bootstrap.
- [ ] Create Bruno API Collection for Product Module.

## Technical Details
- **Database Sequence**: Create `customer_levels` before `products` and `price_tiers` to avoid FK errors.
- **Auto Default Variant**: Creating a product automatically creates a variant with `is_default: true` and `additional_cost: 0`.
