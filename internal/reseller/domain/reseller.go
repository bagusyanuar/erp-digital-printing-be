package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	productDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Reseller struct {
	ID              uuid.UUID                    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerLevelID *uuid.UUID                   `gorm:"type:uuid" json:"customer_level_id"`
	CustomerLevel   *productDomain.CustomerLevel `gorm:"foreignKey:CustomerLevelID" json:"customer_level,omitempty"`
	Name            string                       `gorm:"not null" json:"name"`
	Email           *string                      `gorm:"unique" json:"email"`
	Phone           string                       `json:"phone"`
	Address         string                       `json:"address"`
	CreditLimit     float64                      `gorm:"default:0" json:"credit_limit"`
	OutstandingDebt float64                      `gorm:"->" json:"outstanding_debt"` // read-only field via subquery
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	DeletedAt       gorm.DeletedAt               `gorm:"index" json:"deleted_at,omitempty"`
}

func (r *Reseller) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type ResellerRepository interface {
	Create(ctx context.Context, reseller *Reseller) error
	FindByID(ctx context.Context, id uuid.UUID) (*Reseller, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Reseller, int64, error)
	Update(ctx context.Context, reseller *Reseller) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ResellerUsecase interface {
	Create(ctx context.Context, reseller *Reseller) error
	FindByID(ctx context.Context, id uuid.UUID) (*Reseller, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Reseller, int64, error)
	Update(ctx context.Context, reseller *Reseller) error
	Delete(ctx context.Context, id uuid.UUID) error
}
