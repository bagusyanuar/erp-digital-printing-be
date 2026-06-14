# Phase 14: Expense Module

## Goals
Implement the `expense_categories` and `expenses` tables to track business costs (operational vs production), integrate them directly with `cash_accounts` running balances, and generate analytical reports for the owner.

## Tasks

### 1. Database Schema & Migration
- [x] Create migration files for the new tables: `migrations/xxxxxx_create_expenses_tables.up.sql` and `migrations/xxxxxx_create_expenses_tables.down.sql`.
- [x] Table structure matches `docs/databases/expense.schema.dbml` for `expense_categories` and `expenses`.
- [x] Add relational foreign key constraint to `categories` (product categories) and `users`.

### 2. Domain Model & Repository
- [x] Create `internal/expense/domain/expense.go` to model `ExpenseCategory` and `Expense` structs.
- [x] Implement query/mutator interfaces for repository.
- [x] Implement SQL queries with transaction support for inserting expenses, updating, and fetching analytical data.

### 3. Usecase & Cash Accounts Integration
- [x] Implement logic to create an expense category:
  - If group is `PRODUCTION`, validate that `product_category_id` is present and valid.
  - If group is `OPERATIONAL`, ensure `product_category_id` is stored as `NULL`.
- [x] Implement transaction-safe logic for saving an expense:
  - Acquire pessimistic lock on the target `cash_accounts` (e.g. cash, transfer).
  - Deduct the balance of the target `cash_accounts`.
  - Create the `expenses` record.
  - Insert a `CREDIT` record in `cash_flows` linking `reference_type: 'EXPENSE'` and `reference_id: expense.id`.
- [x] Implement cancellation/deletion logic:
  - Soft-delete or hard-delete the expense.
  - Reverse the cash account balance deduction (add back the amount with lock).
  - Remove/soft-delete the linked `cash_flows` record.

### 4. Handler & Routes
- [x] **Expense Categories Routes:**
  - `POST /api/v1/expense-categories`
  - `GET /api/v1/expense-categories`
  - `PUT /api/v1/expense-categories/:id`
  - `DELETE /api/v1/expense-categories/:id`
- [x] **Expenses Routes:**
  - `POST /api/v1/expenses`
  - `GET /api/v1/expenses`
  - `DELETE /api/v1/expenses/:id`
- [x] **Analytics Routes:**
  - `GET /api/v1/expenses/analytics/summary`
  - `GET /api/v1/expenses/analytics/by-product-category`

## Technical Details
- Use `any` instead of `interface{}`.
- Relational joins, locks, and updates on `cash_accounts`, `cash_flows`, and `expenses` must be executed under a single database transaction.
