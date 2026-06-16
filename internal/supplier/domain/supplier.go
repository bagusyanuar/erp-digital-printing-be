package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Supplier struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	ContactName string         `gorm:"type:varchar(255)" json:"contact_name"`
	Phone       string         `gorm:"type:varchar(50)" json:"phone"`
	Email       string         `gorm:"type:varchar(100)" json:"email"`
	Address     string         `gorm:"type:text" json:"address"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type SupplierRepository interface {
	Create(ctx context.Context, supplier *Supplier) error
	FindByID(ctx context.Context, id uuid.UUID) (*Supplier, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Supplier, int64, error)
	Update(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByName(ctx context.Context, name string) (*Supplier, error)
}

type SupplierUsecase interface {
	Create(ctx context.Context, supplier *Supplier) error
	FindByID(ctx context.Context, id uuid.UUID) (*Supplier, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Supplier, int64, error)
	Update(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
}
