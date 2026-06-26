package dto

import "github.com/google/uuid"

type CreateFundTransferReq struct {
	FromAccount string  `json:"from_account" validate:"required"`
	ToAccount   string  `json:"to_account" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Notes       string  `json:"notes"`
}

type FundTransferRes struct {
	ID           uuid.UUID `json:"id"`
	TransferDate string    `json:"transfer_date"`
	FromAccount  string    `json:"from_account"`
	ToAccount    string    `json:"to_account"`
	Amount       float64   `json:"amount"`
	Notes        *string   `json:"notes,omitempty"`
	CashierName  string    `json:"cashier_name"`
	CreatedAt    string    `json:"created_at"`
}
