package domain

import (
	"context"
	"time"

	categoryDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/category/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UOM Constants
const (
	UomPcs   = "pcs"
	UomM2    = "m2"
	UomMLari = "m_lari"
	UomBox   = "box"
)

// Product model
type Product struct {
	ID         uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CategoryID uuid.UUID               `gorm:"type:uuid;not null" json:"category_id"`
	Category   categoryDomain.Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Name       string                  `gorm:"type:varchar(255);not null" json:"name"`
	SKU        string                  `gorm:"type:varchar(100);unique;not null" json:"sku"`
	UOM        string                  `gorm:"type:varchar(50);not null" json:"uom"`
	BasePrice  float64                 `gorm:"type:decimal(15,2);not null;default:0" json:"base_price"`
	Variants   []ProductVariant        `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
	DeletedAt  gorm.DeletedAt          `gorm:"index" json:"deleted_at,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ProductVariant model
type ProductVariant struct {
	ID              uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID       uuid.UUID               `gorm:"type:uuid;not null" json:"product_id"`
	VariantName     string                  `gorm:"type:varchar(255);not null" json:"variant_name"`
	AdditionalCost  float64                 `gorm:"type:decimal(15,2);not null;default:0" json:"additional_cost"`
	IsDefault       bool                    `gorm:"type:boolean;not null;default:false" json:"is_default"`
	AttributeValues []ProductAttributeValue `gorm:"foreignKey:ProductVariantID" json:"attribute_values,omitempty"`
	PriceTiers      []PriceTier             `gorm:"foreignKey:ProductVariantID" json:"price_tiers,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	DeletedAt       gorm.DeletedAt          `gorm:"index" json:"deleted_at,omitempty"`
}

func (pv *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}
	return nil
}

// Attribute model (EAV definition)
type Attribute struct {
	ID        uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string            `gorm:"type:varchar(255);not null" json:"name"`
	Code      string            `gorm:"type:varchar(100);unique;not null" json:"code"`
	ValueType string            `gorm:"type:varchar(50);not null" json:"value_type"` // text, number, boolean, options
	Options   []AttributeOption `gorm:"foreignKey:AttributeID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
}

func (a *Attribute) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AttributeOption model
type AttributeOption struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AttributeID uuid.UUID      `gorm:"type:uuid;not null" json:"attribute_id"`
	Value       string         `gorm:"type:varchar(255);not null" json:"value"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ao *AttributeOption) BeforeCreate(tx *gorm.DB) error {
	if ao.ID == uuid.Nil {
		ao.ID = uuid.New()
	}
	return nil
}

// ProductAttributeValue model (EAV value link)
type ProductAttributeValue struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductVariantID uuid.UUID      `gorm:"type:uuid;not null" json:"product_variant_id"`
	AttributeID      uuid.UUID      `gorm:"type:uuid;not null" json:"attribute_id"`
	Attribute        Attribute      `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
	Value            string         `gorm:"type:text;not null" json:"value"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (pav *ProductAttributeValue) BeforeCreate(tx *gorm.DB) error {
	if pav.ID == uuid.Nil {
		pav.ID = uuid.New()
	}
	return nil
}

// PriceTier model
type PriceTier struct {
	ID               uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductVariantID uuid.UUID     `gorm:"type:uuid;not null" json:"product_variant_id"`
	CustomerLevelID  uuid.UUID     `gorm:"type:uuid;not null" json:"customer_level_id"`
	CustomerLevel    CustomerLevel `gorm:"foreignKey:CustomerLevelID" json:"customer_level,omitempty"`
	MinQty           int           `gorm:"type:int;not null" json:"min_qty"`
	MaxQty           *int          `gorm:"type:int" json:"max_qty,omitempty"`
	PricePerUnit     float64       `gorm:"type:decimal(15,2);not null" json:"price_per_unit"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (pt *PriceTier) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return nil
}

// Interfaces
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string, categoryID *uuid.UUID) ([]Product, int64, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateVariant(ctx context.Context, variant *ProductVariant) error
}

type ProductUsecase interface {
	Create(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string, categoryID *uuid.UUID) ([]Product, int64, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateVariant(ctx context.Context, variant *ProductVariant) error
}

type AttributeRepository interface {
	Create(ctx context.Context, attribute *Attribute) error
	FindByID(ctx context.Context, id uuid.UUID) (*Attribute, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Attribute, int64, error)
	Update(ctx context.Context, attribute *Attribute) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AttributeUsecase interface {
	Create(ctx context.Context, attribute *Attribute) error
	FindByID(ctx context.Context, id uuid.UUID) (*Attribute, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Attribute, int64, error)
	Update(ctx context.Context, attribute *Attribute) error
	Delete(ctx context.Context, id uuid.UUID) error
}

