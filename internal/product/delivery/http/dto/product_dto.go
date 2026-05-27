package dto

import "github.com/google/uuid"

type CreateProductReq struct {
	CategoryID uuid.UUID `json:"category_id" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	SKU        string    `json:"sku" validate:"required"`
	UOM        string    `json:"uom" validate:"required,oneof=pcs m2 m_lari box"`
	BasePrice  float64   `json:"base_price" validate:"required,numeric,gte=0"`
}

type UpdateProductReq struct {
	CategoryID uuid.UUID `json:"category_id" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	SKU        string    `json:"sku" validate:"required"`
	UOM        string    `json:"uom" validate:"required,oneof=pcs m2 m_lari box"`
	BasePrice  float64   `json:"base_price" validate:"required,numeric,gte=0"`
}

type AttributeValueReq struct {
	AttributeID uuid.UUID `json:"attribute_id" validate:"required"`
	Value       string    `json:"value" validate:"required"`
}

type PriceTierReq struct {
	CustomerLevelID uuid.UUID `json:"customer_level_id" validate:"required"`
	MinQty          int       `json:"min_qty" validate:"required,gt=0"`
	MaxQty          *int      `json:"max_qty" validate:"omitempty,gt=0"`
	PricePerUnit    float64   `json:"price_per_unit" validate:"required,numeric,gte=0"`
}

type CreateVariantReq struct {
	VariantName    string              `json:"variant_name" validate:"required"`
	AdditionalCost float64             `json:"additional_cost" validate:"numeric,gte=0"`
	Attributes     []AttributeValueReq `json:"attributes" validate:"dive"`
	PriceTiers     []PriceTierReq      `json:"price_tiers" validate:"dive"`
}

type AttributeValueRes struct {
	ID          uuid.UUID `json:"id"`
	AttributeID uuid.UUID `json:"attribute_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	ValueType   string    `json:"value_type"`
	Value       string    `json:"value"`
}

type PriceTierRes struct {
	ID                uuid.UUID  `json:"id"`
	CustomerLevelID   uuid.UUID  `json:"customer_level_id"`
	CustomerLevelName string     `json:"customer_level_name"`
	MinQty            int        `json:"min_qty"`
	MaxQty            *int       `json:"max_qty,omitempty"`
	PricePerUnit      float64    `json:"price_per_unit"`
}

type ProductVariantRes struct {
	ID              uuid.UUID           `json:"id"`
	VariantName     string              `json:"variant_name"`
	AdditionalCost  float64             `json:"additional_cost"`
	IsDefault       bool                `json:"is_default"`
	AttributeValues []AttributeValueRes `json:"attribute_values,omitempty"`
	PriceTiers      []PriceTierRes      `json:"price_tiers,omitempty"`
}

type ProductRes struct {
	ID           uuid.UUID           `json:"id"`
	CategoryID   uuid.UUID           `json:"category_id"`
	CategoryName string              `json:"category_name,omitempty"`
	Name         string              `json:"name"`
	SKU          string              `json:"sku"`
	UOM          string              `json:"uom"`
	BasePrice    float64             `json:"base_price"`
	Variants     []ProductVariantRes `json:"variants,omitempty"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
}
