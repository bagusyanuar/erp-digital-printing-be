package domain

import (
	"context"
	"time"

	categoryDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/category/domain"
	supplierDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GroupProduction  = "PRODUCTION"
	GroupOperational = "OPERATIONAL"

	StatusPaid    = "PAID"
	StatusPartial = "PARTIAL"
	StatusUnpaid  = "UNPAID"
)

type ExpenseCategory struct {
	ID                uuid.UUID                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name              string                   `gorm:"type:varchar(255);not null" json:"name"`
	Group             string                   `gorm:"type:varchar(50);not null" json:"group"`
	ProductCategoryID *uuid.UUID               `gorm:"type:uuid" json:"product_category_id"`
	ProductCategory   *categoryDomain.Category `gorm:"foreignKey:ProductCategoryID" json:"product_category"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	DeletedAt         gorm.DeletedAt           `gorm:"index" json:"deleted_at,omitempty"`
}

func (ec *ExpenseCategory) BeforeCreate(tx *gorm.DB) error {
	if ec.ID == uuid.Nil {
		ec.ID = uuid.New()
	}
	return nil
}

type Expense struct {
	ID            uuid.UUID                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseNumber string                   `gorm:"type:varchar(100);uniqueIndex;not null" json:"expense_number"`
	InvoiceNumber *string                  `gorm:"type:varchar(100)" json:"invoice_number"`
	SupplierID    *uuid.UUID               `gorm:"type:uuid" json:"supplier_id"`
	Supplier      *supplierDomain.Supplier `gorm:"foreignKey:SupplierID" json:"supplier"`
	VendorName    string                   `gorm:"type:varchar(255);not null" json:"vendor_name"`
	Amount        float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	Status        string                   `gorm:"type:varchar(50);not null;default:'PAID'" json:"status"`
	ExpenseDate   time.Time                `gorm:"type:timestamp;default:now()" json:"expense_date"`
	Description   *string                  `gorm:"type:text" json:"description"`
	CashierID     uuid.UUID                `gorm:"type:uuid;not null" json:"cashier_id"`
	Discount      float64                  `gorm:"-" json:"discount,omitempty"`
	
	// Relations
	Items         []ExpenseItem            `gorm:"foreignKey:ExpenseID" json:"items"`
	Payments      []ExpensePayment         `gorm:"foreignKey:ExpenseID" json:"payments"`
	
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	DeletedAt     gorm.DeletedAt           `gorm:"index" json:"deleted_at,omitempty"`
}

func (e *Expense) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.ExpenseDate.IsZero() {
		e.ExpenseDate = time.Now()
	}
	return nil
}

type ExpenseItem struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseID         uuid.UUID       `gorm:"type:uuid;not null" json:"expense_id"`
	ExpenseCategoryID uuid.UUID       `gorm:"type:uuid;not null" json:"expense_category_id"`
	ExpenseCategory   ExpenseCategory `gorm:"foreignKey:ExpenseCategoryID" json:"expense_category"`
	Description       *string         `gorm:"type:varchar(255)" json:"description"`
	Amount            float64         `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

func (ei *ExpenseItem) BeforeCreate(tx *gorm.DB) error {
	if ei.ID == uuid.Nil {
		ei.ID = uuid.New()
	}
	return nil
}

type ExpensePayment struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseID     uuid.UUID      `gorm:"type:uuid;not null" json:"expense_id"`
	Amount        float64        `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	PaymentDate   time.Time      `gorm:"type:timestamp;default:now()" json:"payment_date"`
	PaymentMethod string         `gorm:"type:varchar(50);not null" json:"payment_method"`
	CashierID     uuid.UUID      `gorm:"type:uuid;not null" json:"cashier_id"`
	
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ep *ExpensePayment) BeforeCreate(tx *gorm.DB) error {
	if ep.ID == uuid.Nil {
		ep.ID = uuid.New()
	}
	if ep.PaymentDate.IsZero() {
		ep.PaymentDate = time.Now()
	}
	return nil
}

type ExpenseFilter struct {
	StartDate  *time.Time
	EndDate    *time.Time
	Group      string // "PRODUCTION" or "OPERATIONAL"
	CategoryID *uuid.UUID
	Search     string
	Status     string
	Page       int
	Limit      int
}

type ExpenseSummaryRes struct {
	TotalProduction  float64 `json:"total_production"`
	TotalOperational float64 `json:"total_operational"`
	TotalExpense     float64 `json:"total_expense"`
}

type ExpenseWidgetsRes struct {
	TotalExpense      float64 `json:"total_expense"`
	RemainingDebt     float64 `json:"remaining_debt"`
	TransactionVolume int64   `json:"transaction_volume"`
}

type ExpenseByProductCategoryRes struct {
	ProductCategoryID   *uuid.UUID `json:"product_category_id"`
	ProductCategoryName string     `json:"product_category_name"`
	TotalAmount         float64    `json:"total_amount"`
}

type ExpenseRepository interface {
	CreateCategory(ctx context.Context, category *ExpenseCategory) error
	FindCategoryByID(ctx context.Context, id uuid.UUID) (*ExpenseCategory, error)
	FindAllCategories(ctx context.Context, group string) ([]ExpenseCategory, error)
	UpdateCategory(ctx context.Context, category *ExpenseCategory) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	HasAssociatedExpenses(ctx context.Context, categoryID uuid.UUID) (bool, error)

	CreateExpenseTx(ctx context.Context, tx *gorm.DB, expense *Expense) error
	CreateExpenseItemsTx(ctx context.Context, tx *gorm.DB, items []ExpenseItem) error
	CreateExpensePaymentTx(ctx context.Context, tx *gorm.DB, payment *ExpensePayment) error
	UpdateExpenseStatusTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, status string) error
	FindExpenseByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	FindExpenseByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*Expense, error)
	FindAllExpenses(ctx context.Context, filter ExpenseFilter) ([]Expense, int64, error)
	DeleteExpenseTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error

	GetSummary(ctx context.Context, startDate *time.Time, endDate *time.Time) (*ExpenseSummaryRes, error)
	GetByProductCategory(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]ExpenseByProductCategoryRes, error)
	GetWidgets(ctx context.Context, filter ExpenseFilter) (*ExpenseWidgetsRes, error)

	GetDB() *gorm.DB
}

type ExpenseUsecase interface {
	CreateCategory(ctx context.Context, category *ExpenseCategory) error
	FindCategoryByID(ctx context.Context, id uuid.UUID) (*ExpenseCategory, error)
	FindAllCategories(ctx context.Context, group string) ([]ExpenseCategory, error)
	UpdateCategory(ctx context.Context, category *ExpenseCategory) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	CreateExpense(ctx context.Context, expense *Expense) error
	PayInstallment(ctx context.Context, expenseID uuid.UUID, cashierID uuid.UUID, payments []ExpensePayment) error
	FindExpenseByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	FindAllExpenses(ctx context.Context, filter ExpenseFilter) ([]Expense, int64, error)
	DeleteExpense(ctx context.Context, id uuid.UUID) error

	GetSummary(ctx context.Context, startDate *time.Time, endDate *time.Time) (*ExpenseSummaryRes, error)
	GetByProductCategory(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]ExpenseByProductCategoryRes, error)
	GetWidgets(ctx context.Context, filter ExpenseFilter) (*ExpenseWidgetsRes, error)
}
