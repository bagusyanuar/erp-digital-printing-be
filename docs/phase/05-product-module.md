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
- [x] Implement Product Repository (GORM) with preloading of variants and EAV attribute values.
- [x] Implement Usecase logic for Product management (creating default variant upon product creation).
- [x] Create HTTP Handlers and DTOs for Products & Variants.
- [x] Register module in DI Container & Router bootstrap.
- [x] Create Bruno API Collection for Product Module.

## Technical Details
- **Database Sequence**: Create `customer_levels` before `products` and `price_tiers` to avoid FK errors.
- **Auto Default Variant**: Creating a product automatically creates a variant with `is_default: true` and `additional_cost: 0`.

## API Endpoint Specifications

### 1. Create Main Product (Stepper Step 1)
- **Method & URL**: `POST /products`
- **Request Body (JSON)**:
  ```json
  {
    "name": "string (required)",
    "sku": "string (required, unique)",
    "uom": "string (required, oneof=pcs m2 m_lari box)",
    "base_price": "float64 (required, >=0)",
    "category_id": "uuid (required)"
  }
  ```
- **Response**: Mengembalikan objek product yang berhasil dibuat, lengkap dengan **Default Variant** (ID & details) yang otomatis dibuat oleh backend.

### 2. Create Variant with EAV & Price Tiers (Stepper Step 2 & 3 Gabungan)
- **Method & URL**: `POST /products/:product_id/variants`
- **Request Body (JSON)**:
  ```json
  {
    "variant_name": "string (required)",
    "additional_cost": "float64 (required, default 0)",
    "attributes": [
      {
        "attribute_id": "uuid (required)",
        "value": "string (required)"
      }
    ],
    "price_tiers": [
      {
        "customer_level_id": "uuid (required)",
        "min_qty": "int (required)",
        "max_qty": "int (nullable)",
        "price_per_unit": "float64 (required)"
      }
    ]
  }
  ```
- **Response**: Mengembalikan objek `ProductVariant` baru beserta list `attributes` dan `price_tiers` yang terhubung.

