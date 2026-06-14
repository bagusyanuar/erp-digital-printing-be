package domain

import (
	"context"
	"time"

	categoryDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/category/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GroupProduction  = "PRODUCTION"
	GroupOperational = "OPERATIONAL"
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
	ID                uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExpenseCategoryID uuid.UUID        `gorm:"type:uuid;not null" json:"expense_category_id"`
	ExpenseCategory   *ExpenseCategory `gorm:"foreignKey:ExpenseCategoryID" json:"expense_category"`
	Amount            float64          `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	ExpenseDate       time.Time        `gorm:"type:timestamp;default:now()" json:"expense_date"`
	PaymentMethod     string           `gorm:"type:varchar(50);not null" json:"payment_method"`
	Description       *string          `gorm:"type:text" json:"description"`
	CashierID         uuid.UUID        `gorm:"type:uuid;not null" json:"cashier_id"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
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

type ExpenseFilter struct {
	StartDate  *time.Time
	EndDate    *time.Time
	Group      string // "PRODUCTION" or "OPERATIONAL"
	CategoryID *uuid.UUID
	Search     string
	Page       int
	Limit      int
}

type ExpenseSummaryRes struct {
	TotalProduction  float64 `json:"total_production"`
	TotalOperational float64 `json:"total_operational"`
	TotalExpense     float64 `json:"total_expense"`
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
	FindExpenseByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	FindExpenseByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*Expense, error)
	FindAllExpenses(ctx context.Context, filter ExpenseFilter) ([]Expense, int64, error)
	DeleteExpenseTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error

	GetSummary(ctx context.Context, startDate *time.Time, endDate *time.Time) (*ExpenseSummaryRes, error)
	GetByProductCategory(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]ExpenseByProductCategoryRes, error)

	GetDB() *gorm.DB
}

type ExpenseUsecase interface {
	CreateCategory(ctx context.Context, category *ExpenseCategory) error
	FindCategoryByID(ctx context.Context, id uuid.UUID) (*ExpenseCategory, error)
	FindAllCategories(ctx context.Context, group string) ([]ExpenseCategory, error)
	UpdateCategory(ctx context.Context, category *ExpenseCategory) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	CreateExpense(ctx context.Context, expense *Expense) error
	FindAllExpenses(ctx context.Context, filter ExpenseFilter) ([]Expense, int64, error)
	DeleteExpense(ctx context.Context, id uuid.UUID) error

	GetSummary(ctx context.Context, startDate *time.Time, endDate *time.Time) (*ExpenseSummaryRes, error)
	GetByProductCategory(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]ExpenseByProductCategoryRes, error)
}
