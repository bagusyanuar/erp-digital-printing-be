package domain

import (
	"context"
	"time"

	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Capital Transaction Types
const (
	CapitalInjection  = "INJECTION"
	CapitalWithdrawal = "WITHDRAWAL"
)

// CapitalTransaction model
type CapitalTransaction struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionDate time.Time        `gorm:"type:timestamp;not null;default:now()" json:"transaction_date"`
	Type            string           `gorm:"type:varchar(50);not null" json:"type"` // 'INJECTION', 'WITHDRAWAL'
	Amount          float64          `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	PaymentMethod   string           `gorm:"type:varchar(50);not null" json:"payment_method"`
	Description     *string          `gorm:"type:text" json:"description,omitempty"`
	CreatedBy       uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`
	Creator         *userDomain.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

func (ct *CapitalTransaction) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	if ct.TransactionDate.IsZero() {
		ct.TransactionDate = time.Now()
	}
	return nil
}

// CapitalFilter holds query parameters for filtering capital transactions
type CapitalFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Type      string
	Search    string
	Page      int
	Limit     int
}

// CapitalRepository interface
type CapitalRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, transaction *CapitalTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*CapitalTransaction, error)
	FindByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*CapitalTransaction, error)
	FindAll(ctx context.Context, filter CapitalFilter) ([]CapitalTransaction, int64, error)
	DeleteTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
}

// CapitalUsecase interface
type CapitalUsecase interface {
	Create(ctx context.Context, creatorID uuid.UUID, txType string, amount float64, paymentMethod string, description string) (*CapitalTransaction, error)
	FindAll(ctx context.Context, filter CapitalFilter) ([]CapitalTransaction, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
