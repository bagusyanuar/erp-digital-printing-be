# Phase 10: Split Payment Implementation

## Goals
Enable the system to process split payments (e.g. paying partially in cash and partially via QRIS) within a single payment transaction.

## Tasks
- [ ] Update `PaymentProcessReq` DTO to accept an array of payments instead of a single `payment_method` and `amount_paid`.
- [ ] Update `OrderRepayReq` DTO to also accept an array of payments.
- [ ] Modify `ProcessPayment` usecase in `order_usecase.go` to iterate over the payments array, calculate the total paid amount, and insert multiple records into the `order_payments` table.
- [ ] Modify `Repay` usecase in `order_usecase.go` similarly.
- [ ] Update API Documentation (Bruno/OpenAPI) to reflect the new request structures.

## Technical Details
- The `order_payments` table is already structured as One-to-Many with `orders`, so no database schema changes are required. We just need to insert multiple rows per transaction.
- Allowed payment methods: `cash`, `qris`, `transfer`, `tempo`.
