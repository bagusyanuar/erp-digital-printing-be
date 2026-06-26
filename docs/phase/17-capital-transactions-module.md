# Phase 17: Capital Transactions Module (Setoran & Penarikan Modal)

## Goals
Implement the `capital_transactions` table to track setoran modal (injections) and penarikan modal prive (withdrawals) by the owner, integrate them with `cash_accounts` running balances and `cash_flows` ledger, and provide APIs to manage these transactions.

## Tasks

### 1. Database Schema & Migration
- [ ] Create migration files for the new table: `migrations/xxxxxx_create_capital_transactions_table.up.sql` and `migrations/xxxxxx_create_capital_transactions_table.down.sql`.
- [ ] Table structure matches `docs/databases/capital.schema.dbml`.
- [ ] Add foreign key constraint `created_by` referencing `users(id)`.

### 2. Domain Model & Repository
- [ ] Create `internal/capital/domain/capital.go` to model `CapitalTransaction` and filter parameters.
- [ ] Implement query/mutator interfaces for repository.
- [ ] Implement GORM repository for querying, inserting, and deleting (soft-delete) capital transactions.

### 3. Usecase & Cash Accounts Integration
- [ ] Implement transaction-safe logic for creating a capital transaction:
  - Acquire pessimistic lock (`SELECT ... FOR UPDATE`) on the target `cash_accounts` (e.g. cash, transfer).
  - For `INJECTION` (Setoran): Add the amount to the balance of the target `cash_accounts`.
  - For `WITHDRAWAL` (Prive): Deduct the amount from the balance of the target `cash_accounts` (ensure sufficient balance).
  - Create the `capital_transactions` record.
  - Insert a record in `cash_flows` linking `reference_type: 'CAPITAL_INJECTION'` or `'CAPITAL_WITHDRAWAL'` and `reference_id: capital_transaction.id` (DEBIT for injection, CREDIT for withdrawal).
- [ ] Implement deletion (cancellation) logic:
  - Reverse the cash account balance adjustment (subtract for injection, add back for withdrawal with lock).
  - Soft-delete the `capital_transactions` record.
  - Soft-delete/remove the linked `cash_flows` record.

### 4. Handler & Routes
- [ ] **Capital Transaction Routes:**
  - `POST /api/v1/capital` (Create setoran/penarikan modal)
  - `GET /api/v1/capital` (Get history of capital transactions with pagination and type filtering)
  - `DELETE /api/v1/capital/:id` (Cancel/delete capital transaction)

## Technical Details
- Use `any` instead of `interface{}`.
- Mutator operations on `cash_accounts`, `cash_flows`, and `capital_transactions` must be executed under a single database transaction.
