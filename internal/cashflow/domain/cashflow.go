package domain

import (
	"context"
	"time"

	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cash Flow Types
const (
	TypeDebit  = "DEBIT"
	TypeCredit = "CREDIT"
)

// Reference Types
const (
	RefOrderPayment = "ORDER_PAYMENT"
	RefRefund       = "REFUND"
	RefCapital      = "CAPITAL"
	RefAdjustment   = "ADJUSTMENT"
)

// CashFlow model
type CashFlow struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionDate time.Time        `gorm:"type:timestamp;default:now()" json:"transaction_date"`
	ReferenceType   string           `gorm:"type:varchar(50);not null" json:"reference_type"`
	ReferenceID     *uuid.UUID       `gorm:"type:uuid" json:"reference_id,omitempty"`
	Type            string           `gorm:"type:varchar(10);not null" json:"type"`
	Amount          float64          `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	PaymentMethod   string           `gorm:"type:varchar(50);not null" json:"payment_method"`
	Description     *string          `gorm:"type:text" json:"description,omitempty"`
	CustomerName    *string          `gorm:"type:varchar(255)" json:"customer_name,omitempty"`
	CashierID       uuid.UUID        `gorm:"type:uuid;not null" json:"cashier_id"`
	Cashier         *userDomain.User `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

func (cf *CashFlow) BeforeCreate(tx *gorm.DB) error {
	if cf.ID == uuid.Nil {
		cf.ID = uuid.New()
	}
	if cf.TransactionDate.IsZero() {
		cf.TransactionDate = time.Now()
	}
	return nil
}

// Cash Flow Report Structures
type CashFlowSummary struct {
	TotalDebit  float64 `json:"total_debit"`
	TotalCredit float64 `json:"total_credit"`
	NetBalance  float64 `json:"net_balance"`
}

type CashFlowMethodDetail struct {
	Debit   float64 `json:"debit"`
	Credit  float64 `json:"credit"`
	Balance float64 `json:"balance"`
}

type CashFlowTransactionRes struct {
	ID              uuid.UUID  `json:"id"`
	TransactionDate time.Time  `json:"transaction_date"`
	ReferenceType   string     `json:"reference_type"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	Type            string     `json:"type"`
	Amount          float64    `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	Description     *string    `json:"description,omitempty"`
	CustomerName    *string    `json:"customer_name,omitempty"`
	CashierName     string     `json:"cashier_name"`
}

// CashFlowFilter menampung query parameter untuk pencarian dan paginasi
type CashFlowFilter struct {
	StartDate     time.Time
	EndDate       time.Time
	PaymentMethod string
	Type          string
	ReferenceType string
	CashierID     *uuid.UUID
	Search        string
	Page          int
	Limit         int
}

// CashFlowReportRes sudah tidak digunakan jika response dipisah, 
// namun kita definisikan struct baru untuk Summary response nanti jika dibutuhkan.
type CashFlowSummaryRes struct {
	Summary         CashFlowSummary                 `json:"summary"`
	DetailsByMethod map[string]CashFlowMethodDetail `json:"details_by_method"`
}

// CashAccount model
type CashAccount struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:varchar(50);unique;not null" json:"name"`
	Balance   float64   `gorm:"type:decimal(15,2);not null;default:0" json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ca *CashAccount) BeforeCreate(tx *gorm.DB) error {
	if ca.ID == uuid.Nil {
		ca.ID = uuid.New()
	}
	return nil
}

// Interfaces
type CashFlowRepository interface {
	Create(ctx context.Context, cashFlow *CashFlow) error
	CreateTx(ctx context.Context, tx *gorm.DB, cashFlow *CashFlow) error
	FindAll(ctx context.Context, filter CashFlowFilter) ([]CashFlow, int64, error)
	GetSummary(ctx context.Context, filter CashFlowFilter) (*CashFlowSummaryRes, error)
	FindAllAccounts(ctx context.Context) ([]CashAccount, error)
	FindAccountByNameWithLock(ctx context.Context, tx *gorm.DB, name string) (*CashAccount, error)
	UpdateAccount(ctx context.Context, tx *gorm.DB, account *CashAccount) error
}

type CashFlowUsecase interface {
	GetReport(ctx context.Context, filter CashFlowFilter) ([]CashFlowTransactionRes, int64, error)
	GetSummary(ctx context.Context, filter CashFlowFilter) (*CashFlowSummaryRes, error)
	CreateAdjustment(ctx context.Context, cashierID uuid.UUID, amount float64, flowType string, paymentMethod string, description string) (*CashFlow, error)
	FindAllAccounts(ctx context.Context) ([]CashAccount, error)
}

