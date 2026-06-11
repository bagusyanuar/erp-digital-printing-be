# Phase 12: Cash Flow Module

## Goals
Implement the Cash Flow (General Ledger) system to centralize and track all financial mutations (DEBIT for incoming, CREDIT for outgoing) in the ERP.

## Tasks

### 1. Database Schema & Migration
- [ ] Create a new migration file: `migrations/xxxxxx_create_cash_flows_table.up.sql` and `migrations/xxxxxx_create_cash_flows_table.down.sql` (to be run manually).
- [ ] Table structure matches `docs/databases/cash_flow.schema.dbml`.

### 2. Domain Model & Repository
- [ ] Create `internal/cashflow/domain/cashflow.go` for the domain entities and interfaces.
- [ ] Implement `internal/cashflow/repository/cashflow_repository.go` to handle DB operations using GORM.

### 3. Usecase & Integrations
- [ ] Implement `internal/cashflow/usecase/cashflow_usecase.go` to handle the business logic:
  - Generate Cash Flow Reports (Summary, breakdown by payment method, and transaction list).
  - Create manual cash flow adjustments (e.g., surplus/shortage).
- [ ] Integrate into `OrderUsecase`:
  - When a payment is processed (`ProcessPayment` or `ProcessRepayment`), automatically insert a `DEBIT` record into `cash_flows` within the same DB transaction.

### 4. Handler & Routing
- [ ] Implement `internal/cashflow/delivery/http/cashflow_handler.go` to expose:
  - `GET /api/v1/reports/cash-flow?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`
- [ ] Register the new routes in the HTTP server setup.

## Technical Details
- Use `any` for generic interfaces instead of `interface{}`.
- All database mutation updates must run in a GORM transaction when coupled with other modules (e.g., order payment).
