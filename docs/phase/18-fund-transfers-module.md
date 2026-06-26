# Phase 18: Fund Transfers Module (Pemindahan Dana Antar Akun)

## Goals
Implement the `fund_transfers` table to track internal balance movements between cash accounts, integrate with `cash_accounts` running balances and `cash_flows` ledger, and provide APIs to execute and view fund transfer history.

## Tasks

### 1. Database Schema & Migration
- [ ] Create migration files for the new table: `migrations/xxxxxx_create_fund_transfers_table.up.sql` and `migrations/xxxxxx_create_fund_transfers_table.down.sql`.
- [ ] Table structure matches `docs/databases/cash_flow.schema.dbml`.
- [ ] Add foreign key constraints for `from_account_id` and `to_account_id` referencing `cash_accounts(id)`, and `cashier_id` referencing `users(id)`.

### 2. Domain Model & Repository
- [ ] Create `internal/cashflow/domain/fund_transfer.go` to model `FundTransfer` and query filter parameters.
- [ ] Implement GORM repository for querying, inserting, and soft-deleting fund transfers.

### 3. Usecase & Cash Accounts Integration
- [ ] Implement transaction-safe logic for executing a fund transfer:
  - Start database transaction.
  - Acquire pessimistic locks (`SELECT ... FOR UPDATE`) on both origin (`from_account_id`) and destination (`to_account_id`) `cash_accounts`. Sort the IDs before locking to prevent deadlock.
  - Ensure the origin account has sufficient balance.
  - Deduct the amount from the origin account, add the amount to the destination account.
  - Create the `fund_transfers` record.
  - Insert two records in `cash_flows`:
    - Record 1: `CREDIT` for origin account, `reference_type: 'FUND_TRANSFER'`, `reference_id: fund_transfer.id`.
    - Record 2: `DEBIT` for destination account, `reference_type: 'FUND_TRANSFER'`, `reference_id: fund_transfer.id`.
  - Commit transaction.
- [ ] Implement cancellation/deletion logic:
  - Acquire locks on both accounts.
  - Reverse the balances (add back to origin, deduct from destination).
  - Soft-delete the `fund_transfers` record.
  - Soft-delete the associated `cash_flows` records.

### 4. Handler & Routes
- [ ] **Fund Transfer Routes:**
  - `POST /api/v1/fund-transfers` (Execute transfer)
  - `GET /api/v1/fund-transfers` (Get transfer history with pagination and date filters)
  - `DELETE /api/v1/fund-transfers/:id` (Cancel/delete a transfer)

## Technical Details
- Use `any` instead of `interface{}`.
- Mutator operations must run under a single DB transaction.
