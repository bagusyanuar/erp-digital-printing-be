# Phase 13: Cash Accounts Integration (Option C Running Balance)

## Goals
Implement the `cash_accounts` master table to track dynamic payment methods (e.g., Cash, Transfer, QRIS) and maintain a real-time running balance using Pessimistic Locking (`SELECT ... FOR UPDATE`) to prevent race conditions.

## Tasks

### 1. Database Schema & Migration
- [ ] Create a new migration file: `migrations/xxxxxx_create_cash_accounts_table.up.sql` and `migrations/xxxxxx_create_cash_accounts_table.down.sql` (to be run manually).
- [ ] Table structure matches `docs/databases/cash_flow.schema.dbml` for `cash_accounts`.
- [ ] Seed default payment methods: `cash`, `transfer`, `qris` with initial balance of `0.00`.

### 2. Domain Model & Repository Updates
- [ ] Create `internal/cashflow/domain/cash_account.go` (or add to `cashflow.go`) to model `CashAccount`.
- [ ] Implement query methods to fetch cash accounts.
- [ ] Implement `GetAccountByNameWithLock` or similar using GORM pessimistic locking:
  `tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, "name = ?", name)`

### 3. Usecase & Concurrency Integration
- [ ] Modify `CreateAdjustment` to lock the target `CashAccount`, update its balance, and save the transaction.
- [ ] Modify `OrderUsecase` integration: when recording order payments, lock the corresponding `CashAccount` and update its balance in the same database transaction.

### 4. Handler & Routes
- [ ] Add `GET /api/v1/cash-accounts` to expose the active accounts and their current balances for both front-office checkout page and back-office reports.

## Technical Details
- Use `any` instead of `interface{}`.
- DB updates to `cash_accounts` and `cash_flows` must be executed under a database transaction with a row-level write lock.
