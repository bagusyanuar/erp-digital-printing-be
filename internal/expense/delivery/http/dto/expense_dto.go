package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCategoryReq struct {
	Name              string     `json:"name"`
	Group             string     `json:"group"`
	ProductCategoryID *uuid.UUID `json:"product_category_id"`
}

type UpdateCategoryReq struct {
	Name              string     `json:"name"`
	Group             string     `json:"group"`
	ProductCategoryID *uuid.UUID `json:"product_category_id"`
}

type CreateExpenseReq struct {
	ExpenseCategoryID uuid.UUID  `json:"expense_category_id"`
	Amount            float64    `json:"amount"`
	ExpenseDate       *time.Time `json:"expense_date"`
	PaymentMethod     string     `json:"payment_method"`
	Description       *string    `json:"description"`
}
