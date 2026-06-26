package domain

import (
	"context"
	"time"

	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FundTransfer struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransferDate time.Time        `gorm:"type:timestamp;default:now()" json:"transfer_date"`
	FromAccountID uuid.UUID       `gorm:"type:uuid;not null" json:"from_account_id"`
	FromAccount   *CashAccount     `gorm:"foreignKey:FromAccountID" json:"from_account,omitempty"`
	ToAccountID   uuid.UUID       `gorm:"type:uuid;not null" json:"to_account_id"`
	ToAccount     *CashAccount     `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
	Amount       float64          `gorm:"type:decimal(15,2);not null" json:"amount"`
	Notes        *string          `gorm:"type:text" json:"notes,omitempty"`
	CashierID    uuid.UUID        `gorm:"type:uuid;not null" json:"cashier_id"`
	Cashier      *userDomain.User `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

func (ft *FundTransfer) BeforeCreate(tx *gorm.DB) error {
	if ft.ID == uuid.Nil {
		ft.ID = uuid.New()
	}
	if ft.TransferDate.IsZero() {
		ft.TransferDate = time.Now()
	}
	return nil
}

type FundTransferFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	Limit     int
}

type FundTransferRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, transfer *FundTransfer) error
	FindByID(ctx context.Context, id uuid.UUID) (*FundTransfer, error)
	FindAll(ctx context.Context, filter FundTransferFilter) ([]FundTransfer, int64, error)
	DeleteTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
}

type FundTransferUsecase interface {
	Transfer(ctx context.Context, cashierID uuid.UUID, fromAccountName string, toAccountName string, amount float64, notes string) (*FundTransfer, error)
	FindAll(ctx context.Context, filter FundTransferFilter) ([]FundTransfer, int64, error)
	Cancel(ctx context.Context, cashierID uuid.UUID, id uuid.UUID) error
}
