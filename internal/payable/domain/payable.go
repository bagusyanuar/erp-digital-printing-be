package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
)

type VendorPayable struct {
	ID                  string               `json:"id"`
	VendorName          string               `json:"vendor_name"`
	VendorPhone         string               `json:"vendor_phone"`
	VendorCategory      string               `json:"vendor_category,omitempty"`
	TotalExpensesCount  int                  `json:"total_expenses_count"`
	UnpaidExpensesCount int                  `json:"unpaid_expenses_count"`
	TotalAmount         float64              `json:"total_amount"`
	TotalPaid           float64              `json:"total_paid"`
	TotalRemaining      float64              `json:"total_remaining"`
	Expenses            []PayableExpenseItem `json:"expenses"`
}

type PayableExpenseItem struct {
	ID              uuid.UUID            `json:"id"`
	ExpenseNumber   string               `json:"expense_number"`
	ReferenceNumber *string              `json:"reference_number,omitempty"`
	CategoryName    string               `json:"category_name"`
	ExpenseDate     string               `json:"expense_date"`
	TotalAmount     float64              `json:"total_amount"`
	AmountPaid      float64              `json:"amount_paid"`
	RemainingAmount float64              `json:"remaining_amount"`
	PaymentStatus   string               `json:"payment_status"`
	Notes           *string              `json:"notes,omitempty"`
	Payments        []PayablePaymentItem `json:"payments"`
}

type PayablePaymentItem struct {
	ID               uuid.UUID `json:"id"`
	PayableExpenseID uuid.UUID `json:"payable_expense_id"`
	Amount           float64   `json:"amount"`
	PaymentMethod    string    `json:"payment_method"`
	RecordedBy       string    `json:"recorded_by"`
	PaymentDate      string    `json:"payment_date"`
	SourceAccount    *string   `json:"source_account,omitempty"`
	Notes            *string   `json:"notes,omitempty"`
}

type SettledExpensePayable struct {
	ID              uuid.UUID            `json:"id"`
	VendorID        string               `json:"vendor_id"`
	VendorName      string               `json:"vendor_name"`
	VendorPhone     string               `json:"vendor_phone"`
	VendorCategory  string               `json:"vendor_category,omitempty"`
	ExpenseNumber   string               `json:"expense_number"`
	ReferenceNumber *string              `json:"reference_number,omitempty"`
	CategoryName    string               `json:"category_name"`
	ExpenseDate     string               `json:"expense_date"`
	SettledDate     string               `json:"settled_date"`
	TotalAmount     float64              `json:"total_amount"`
	AmountPaid      float64              `json:"amount_paid"`
	PaymentMethod   string               `json:"payment_method"`
	Payments        []PayablePaymentItem `json:"payments"`
	Notes           *string              `json:"notes,omitempty"`
}

type VendorPayableSummary struct {
	TotalActivePayable  float64 `json:"total_active_payable"`
	TotalSettledPayable float64 `json:"total_settled_payable"`
	ActiveVendorCount   int64   `json:"active_vendor_count"`
}

type PayVendorExpenseInput struct {
	ExpenseID uuid.UUID
	CashierID uuid.UUID
	Payments  []PayExpenseItemInput
	Notes     *string
}

type PayExpenseItemInput struct {
	PaymentMethod string
	Amount        float64
	SourceAccount *string
}

type PayableRepository interface {
	FindVendorPayables(ctx context.Context, params request.PaginationParam, search string, status string, startDate *time.Time, endDate *time.Time) ([]VendorPayable, int64, error)
	FindSettledPayables(ctx context.Context, params request.PaginationParam, search string, startDate *time.Time, endDate *time.Time) ([]SettledExpensePayable, int64, error)
	GetSummary(ctx context.Context) (*VendorPayableSummary, error)
	FindVendorPayableByExpenseID(ctx context.Context, expenseID uuid.UUID) (*VendorPayable, error)
}

type PayableUsecase interface {
	GetVendorPayables(ctx context.Context, params request.PaginationParam, search string, status string, startDate *time.Time, endDate *time.Time) ([]VendorPayable, int64, error)
	GetSettledPayables(ctx context.Context, params request.PaginationParam, search string, startDate *time.Time, endDate *time.Time) ([]SettledExpensePayable, int64, error)
	GetSummary(ctx context.Context) (*VendorPayableSummary, error)
	PayExpense(ctx context.Context, input PayVendorExpenseInput) (*VendorPayable, error)
}

