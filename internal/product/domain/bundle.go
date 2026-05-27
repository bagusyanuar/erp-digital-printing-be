package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bundle model
type Bundle struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	SKU       string         `gorm:"type:varchar(100);unique;not null" json:"sku"`
	BasePrice float64        `gorm:"type:decimal(15,2);not null" json:"base_price"`
	Items     []BundleItem   `gorm:"foreignKey:BundleID" json:"items,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (b *Bundle) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BundleItem model
type BundleItem struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BundleID         uuid.UUID      `gorm:"type:uuid;not null" json:"bundle_id"`
	ProductVariantID uuid.UUID      `gorm:"type:uuid;not null" json:"product_variant_id"`
	ProductVariant   ProductVariant `gorm:"foreignKey:ProductVariantID" json:"product_variant,omitempty"`
	Qty              int            `gorm:"type:int;not null" json:"qty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (bi *BundleItem) BeforeCreate(tx *gorm.DB) error {
	if bi.ID == uuid.Nil {
		bi.ID = uuid.New()
	}
	return nil
}

// Interfaces
type BundleRepository interface {
	Create(ctx context.Context, bundle *Bundle) error
	FindByID(ctx context.Context, id uuid.UUID) (*Bundle, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Bundle, int64, error)
	Update(ctx context.Context, bundle *Bundle) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type BundleUsecase interface {
	Create(ctx context.Context, bundle *Bundle) error
	FindByID(ctx context.Context, id uuid.UUID) (*Bundle, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Bundle, int64, error)
	Update(ctx context.Context, bundle *Bundle) error
	Delete(ctx context.Context, id uuid.UUID) error
}
