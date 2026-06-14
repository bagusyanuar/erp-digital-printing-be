package dto

import "github.com/google/uuid"

type CreateAdjustmentReq struct {
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Type          string  `json:"type" validate:"required,oneof=DEBIT CREDIT"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash transfer qris"`
	Description   string  `json:"description" validate:"required"`
}

type CashFlowRes struct {
	ID              uuid.UUID  `json:"id"`
	TransactionDate string     `json:"transaction_date"`
	ReferenceType   string     `json:"reference_type"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	Type            string     `json:"type"`
	Amount          float64    `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	Description     *string    `json:"description,omitempty"`
	InvoiceNumber   *string    `json:"invoice_number,omitempty"`
	CashierID       uuid.UUID  `json:"cashier_id"`
}
