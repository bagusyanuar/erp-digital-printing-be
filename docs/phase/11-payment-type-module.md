# Phase 11: Payment Type Classification

## Goals
Explicitly classify each payment log as a Down Payment, Full Payment, or Repayment rather than relying on array indices in the frontend.

## Tasks
- [ ] Create a new database migration to add a `payment_type` column to the `order_payments` table.
- [ ] Update the `OrderPayment` struct in `internal/order/domain/order.go` to include the `PaymentType` field.
- [ ] Update the `ProcessPayment` and `Repay` usecases to populate the `payment_type` based on the context of the payment.
  * `DOWN_PAYMENT` / `FULL_PAYMENT` for initial payments.
  * `REPAYMENT` for subsequent payments.
- [ ] Expose `PaymentType` in the `OrderPaymentRes` DTO for the frontend to consume.

## Technical Details
- The new column should be a `VARCHAR(50)`.
- Existing records in `order_payments` may need a default value (e.g. `UNKNOWN` or derived from order status) during migration.
