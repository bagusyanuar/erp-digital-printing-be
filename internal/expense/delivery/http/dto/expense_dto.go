package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCategoryReq struct {
	Name              string     `json:"name"`
	Group             string     `json:"group"`
	ProductCategoryID *uuid.UUID `json:"product_category_id"`
	Sequence          int        `json:"sequence,omitempty"`
}

type UpdateCategoryReq struct {
	Name              string     `json:"name"`
	Group             string     `json:"group"`
	ProductCategoryID *uuid.UUID `json:"product_category_id"`
	Sequence          int        `json:"sequence,omitempty"`
}

type CreateExpenseReq struct {
	InvoiceNumber *string          `json:"invoice_number"`
	SupplierID    *uuid.UUID       `json:"supplier_id"`
	VendorName    string           `json:"vendor_name"`
	ExpenseDate   *time.Time       `json:"expense_date"`
	Description   *string          `json:"description"`
	Discount      float64          `json:"discount"`
	Items         []ExpenseItemReq `json:"items"`
	Payments      []PaymentReq     `json:"payments"`
}

type ExpenseItemReq struct {
	ExpenseCategoryID uuid.UUID `json:"expense_category_id"`
	Description       *string   `json:"description"`
	Amount            float64   `json:"amount"`
}

type PaymentReq struct {
	Amount        float64    `json:"amount"`
	PaymentMethod string     `json:"payment_method"`
	PaymentDate   *time.Time `json:"payment_date"`
}

type PayInstallmentReq struct {
	Payments []PaymentReq `json:"payments"`
}

type CategoryExpenseItemRes struct {
	ID                  uuid.UUID  `json:"id"`
	CategoryID          *uuid.UUID `json:"category_id"`
	CategoryName        string     `json:"category_name"`
	ExpenseCategoryID   uuid.UUID  `json:"expense_category_id"`
	ExpenseCategoryName string     `json:"expense_category_name"`
	Description         string     `json:"description"`
	TotalExpense        float64    `json:"total_expense"`
	ExpenseCount        int64      `json:"expense_count"`
}
