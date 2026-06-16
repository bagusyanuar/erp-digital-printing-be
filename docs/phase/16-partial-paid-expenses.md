# Phase 16: Partial Paid & Split Payment Expense Module

## Goals
Refactor the Expense module to support split payment (Tunai + Transfer), term payment (termin/cicilan/hutang), and itemized expense entry (header-detail pattern), integrated with registered suppliers and cash flow.

## Tasks

### 1. Database Schema & Migration
- [ ] Create migration files: `migrations/000017_refactor_expenses_tables.up.sql` and `migrations/000017_refactor_expenses_tables.down.sql`.
- [ ] Refactor table `expenses`:
  - Drop column `expense_category_id` and `payment_method`.
  - Add column `expense_number` (VARCHAR, UNIQUE).
  - Add column `invoice_number` (VARCHAR, NULL).
  - Add column `supplier_id` (UUID, FK -> `suppliers.id` ON DELETE SET NULL).
  - Add column `vendor_name` (VARCHAR, NOT NULL, DEFAULT 'Umum').
  - Add column `status` (VARCHAR, NOT NULL, DEFAULT 'PAID').
- [ ] Create table `expense_items` (detail item belanja).
- [ ] Create table `expense_payments` (riwayat split payment/cicilan/termin).
- [ ] Seed default category `'Potongan Pembelian'` into `expense_categories`.

### 2. Domain Model & Repository
- [ ] Refactor `internal/expense/domain/expense.go`:
  - Update `Expense` struct to include relation mappings to `ExpenseItem` and `ExpensePayment`.
  - Define `ExpenseItem` and `ExpensePayment` structs.
- [ ] Update repository interface and SQL queries:
  - Add support for transactions to save header (`expenses`), detail items (`expense_items`), and payments (`expense_payments`).
  - Implement query to retrieve payments by expense ID.
  - Implement query to get total paid amount per expense ID.

### 3. Usecase & Cash Flow Integration
- [ ] Refactor `CreateExpense` logic:
  - Generate automatic `expense_number` (e.g. `EXP/YYYYMMDD/XXXX`).
  - Calculate total billing (`total_belanja`) from `items` (support negative price for 'Potongan Pembelian').
  - Calculate total initial payment (`total_bayar`) from `payments` array.
  - Determine status:
    - `total_bayar == total_belanja` -> `'PAID'`
    - `total_bayar > 0 && total_bayar < total_belanja` -> `'PARTIAL'`
    - `total_bayar == 0` -> `'UNPAID'`
  - Save to database inside a single transaction (`tx`):
    - Create header `expenses` and bulk insert `expense_items`.
    - Loop and insert `expense_payments`.
    - For each payment, pessimistically lock target `cash_accounts`, deduct balance, and create `cash_flows` record with `reference_type = 'EXPENSE_PAYMENT'`.
- [ ] Implement `PayInstallment` (Repayment) logic:
  - Lock expense header.
  - Validate total accumulated payments + new payments does not exceed `total_belanja`.
  - Insert new `expense_payments` record.
  - Update cash account balances and record to `cash_flows` for each payment.
  - Calculate and update the `status` of the `expenses` (change `'PARTIAL'` or `'UNPAID'` to `'PAID'` if fully settled).

### 4. Handler & Routes
- [ ] **Expenses Routes:**
  - `POST /api/v1/expenses` (Request body supports items list & payments array).
  - `POST /api/v1/expenses/:id/payments` (Add payment installment/split payment cicilan).
  - `GET /api/v1/expenses` (Includes filters by status, vendor name, etc., and supports eager loading for items).
  - `GET /api/v1/expenses/:id` (Get expense detail with items and payment history).

## Technical Details
- Use `any` instead of `interface{}`.
- Logging using Zap logger.
- Relational joins, locks, and updates on cash accounts/flows under single database transaction.
- GORM preload to fetch header and items/payments relation.
