package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
)

type CustomerReceivable struct {
	ID                  string                  `json:"id"`
	CustomerName        string                  `json:"customer_name"`
	CustomerPhone       string                  `json:"customer_phone"`
	CustomerType        string                  `json:"customer_type"`
	TotalInvoicesCount  int                     `json:"total_invoices_count"`
	UnpaidInvoicesCount int                     `json:"unpaid_invoices_count"`
	TotalAmount         float64                 `json:"total_amount"`
	TotalPaid           float64                 `json:"total_paid"`
	TotalRemaining      float64                 `json:"total_remaining"`
	Invoices            []ReceivableInvoiceItem `json:"invoices"`
}

type ReceivableInvoiceItem struct {
	ID              uuid.UUID               `json:"id"`
	InvoiceNumber   string                  `json:"invoice_number"`
	JobNumber       string                  `json:"job_number"`
	OrderDate       string                  `json:"order_date"`
	TotalAmount     float64                 `json:"total_amount"`
	AmountPaid      float64                 `json:"amount_paid"`
	RemainingAmount float64                 `json:"remaining_amount"`
	PaymentStatus   string                  `json:"payment_status"`
	Notes           *string                 `json:"notes,omitempty"`
	Payments        []ReceivablePaymentItem `json:"payments"`
}

type ReceivablePaymentItem struct {
	ID                  uuid.UUID `json:"id"`
	ReceivableInvoiceID uuid.UUID `json:"receivable_invoice_id"`
	Amount              float64   `json:"amount"`
	PaymentMethod       string    `json:"payment_method"`
	CashierName         string    `json:"cashier_name"`
	PaymentDate         string    `json:"payment_date"`
}

type SettledInvoiceReceivable struct {
	ID            uuid.UUID               `json:"id"`
	CustomerID    string                  `json:"customer_id"`
	CustomerName  string                  `json:"customer_name"`
	CustomerPhone string                  `json:"customer_phone"`
	CustomerType  string                  `json:"customer_type"`
	InvoiceNumber string                  `json:"invoice_number"`
	JobNumber     string                  `json:"job_number"`
	OrderDate     string                  `json:"order_date"`
	SettledDate   string                  `json:"settled_date"`
	TotalAmount   float64                 `json:"total_amount"`
	AmountPaid    float64                 `json:"amount_paid"`
	PaymentMethod string                  `json:"payment_method"`
	Payments      []ReceivablePaymentItem `json:"payments"`
	Notes         *string                 `json:"notes,omitempty"`
}

type CustomerReceivableSummary struct {
	TotalActiveReceivable  float64 `json:"total_active_receivable"`
	TotalSettledReceivable float64 `json:"total_settled_receivable"`
	ActiveCustomerCount    int64   `json:"active_customer_count"`
}

type ReceivableRepository interface {
	FindCustomerReceivables(ctx context.Context, params request.PaginationParam, search string, status string, startDate *time.Time, endDate *time.Time) ([]CustomerReceivable, int64, error)
	FindSettledInvoices(ctx context.Context, params request.PaginationParam, search string, startDate *time.Time, endDate *time.Time) ([]SettledInvoiceReceivable, int64, error)
	GetSummary(ctx context.Context) (*CustomerReceivableSummary, error)
}

type ReceivableUsecase interface {
	GetCustomerReceivables(ctx context.Context, params request.PaginationParam, search string, status string, startDate *time.Time, endDate *time.Time) ([]CustomerReceivable, int64, error)
	GetSettledInvoices(ctx context.Context, params request.PaginationParam, search string, startDate *time.Time, endDate *time.Time) ([]SettledInvoiceReceivable, int64, error)
	GetSummary(ctx context.Context) (*CustomerReceivableSummary, error)
}

