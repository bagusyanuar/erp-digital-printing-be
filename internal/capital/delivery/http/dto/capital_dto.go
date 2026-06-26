package dto

import "github.com/google/uuid"

type CreateCapitalReq struct {
	Type          string  `json:"type" validate:"required,oneof=INJECTION WITHDRAWAL"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" validate:"required"`
	Description   string  `json:"description"`
}

type CapitalTransactionRes struct {
	ID              uuid.UUID  `json:"id"`
	TransactionDate string     `json:"transaction_date"`
	Type            string     `json:"type"`
	Amount          float64    `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	Description     *string    `json:"description,omitempty"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	CreatorName     string     `json:"creator_name"`
	CreatedAt       string     `json:"created_at"`
}
