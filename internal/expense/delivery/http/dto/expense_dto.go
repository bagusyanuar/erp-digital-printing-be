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
	Qty               int       `json:"qty"`
	Price             float64   `json:"price"`
}

type PaymentReq struct {
	Amount        float64    `json:"amount"`
	PaymentMethod string     `json:"payment_method"`
	PaymentDate   *time.Time `json:"payment_date"`
}

type PayInstallmentReq struct {
	Payments []PaymentReq `json:"payments"`
}
