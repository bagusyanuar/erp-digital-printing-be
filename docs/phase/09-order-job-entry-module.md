# Phase 9: Order & Job Entry Module Implementation

## Goals
Implement the order transaction system, job entry tracking, and workflow transitions (from Designer Draft to Cashier Queue).

## Tasks

### 1. Database & Migrations
- [ ] Create database migration for order and job entry module:
  * `finishings` (Master data for laminasi, mata ayam, etc.)
  * `orders` (Main transaction & job headers)
  * `order_items` (Individual job items)
  * `order_item_finishings` (Pivot table for item finishing choices)
- [ ] Create database seeder for default `finishings` (e.g. Mata Ayam, Laminasi Glossy, Laminasi Doff).
- [ ] Execute migration manually (`make migrate-up`).

### 2. Domain & Core Entities (Go)
- [ ] Define Go domain structs with GORM tags:
  * `Order`
  * `OrderItem`
  * `Finishing`
- [ ] Implement `BeforeCreate` GORM hook for UUID generation in all entities.
- [ ] Define Repository & Usecase Interfaces for Orders & Finishings.

### 3. Job Entry & Order CRUD (Go Backend)
- [ ] Implement `OrderRepository` (GORM) supporting:
  * Creating draft orders (`status = DRAFT` / `PENDING_PAYMENT`).
  * Querying cashier queues (`status = PENDING_PAYMENT`).
  * Preloading nested `order_items`, `order_item_finishings`, and `reseller`/`designer` details.
- [ ] Implement Usecase logic for Order management:
  * Validation rules for UOM (`m2` / `m_lari` dimension checks).
  * Auto-generation of `job_number` using format `JOB/YYYYMMDD/XXXX`.
  * Auto-generation of `invoice_number` using format `INV/YYYYMMDD/XXXX` only when paid.
- [ ] Create HTTP Handlers and DTOs for Orders.
- [ ] Register module in DI Container & Router bootstrap.
- [ ] Create Bruno/Postman API Collection for testing.

---

## Technical Details

- **Job Number Generator**: Uses format `JOB/YYYYMMDD/XXXX` with an atomic daily counter sequence in PostgreSQL or Redis.
- **Invoice Number Generator**: Uses format `INV/YYYYMMDD/XXXX` with an atomic daily counter sequence, created ONLY upon transition from `PENDING_PAYMENT` to `IN_PRODUCTION`.
- **UOM Validations**:
  * `uom == "m2"`: Requires `length_cm > 0` and `width_cm > 0`.
  * `uom == "m_lari"`: Requires `length_cm > 0`.

---

## API Endpoint Specifications

### 1. Save Job Entry Draft (Designer Hold)
- **Method & URL**: `POST /api/v1/orders/draft`
- **Role**: Designer
- **Request Body (JSON)**:
  ```json
  {
    "designer_id": "uuid (required)",
    "customer_name": "string (optional)",
    "customer_phone": "string (optional)",
    "notes": "string (optional)",
    "items": [
      {
        "product_variant_id": "uuid (required)",
        "uom": "string (required, oneof=pcs m2 m_lari box)",
        "length_cm": "float64 (optional)",
        "width_cm": "float64 (optional)",
        "quantity": "int (required, >0)",
        "design_file_url": "string (optional)",
        "production_notes": "string (optional)",
        "finishing_ids": ["uuid"]
      }
    ]
  }
  ```
- **Response**: Returns created Order with `status: "DRAFT"`, `job_number`, and `invoice_number: null`.

### 2. Submit Job Entry to Cashier Queue
- **Method & URL**: `POST /api/v1/orders/submit`
- **Role**: Designer
- **Description**: Immediately locks specifications and sends to cashier queue. (Can also transition an existing `DRAFT` via `PUT /api/v1/orders/:id/submit`).
- **Response**: Returns Order with `status: "PENDING_PAYMENT"`.

### 3. Fetch Cashier Queue
- **Method & URL**: `GET /api/v1/orders/cashier-queue`
- **Role**: Cashier
- **Description**: Returns all orders where `status: "PENDING_PAYMENT"`.

### 4. Process Payment & Release to Production
- **Method & URL**: `POST /api/v1/orders/:id/pay`
- **Role**: Cashier
- **Request Body (JSON)**:
  ```json
  {
    "reseller_id": "uuid (optional)",
    "customer_name": "string (required, to confirm identity)",
    "customer_phone": "string (required)",
    "payment_method": "string (required, e.g. cash, transfer)",
    "payment_type": "string (required, oneof=full dp)",
    "amount_paid": "float64 (required, >0)"
  }
  ```
- **Description**: Performs pricing engine calculations, validates payment amount, generates `invoice_number`, sets `status = "IN_PRODUCTION"`.
