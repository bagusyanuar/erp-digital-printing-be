package dto

import "github.com/google/uuid"

type OrderItemReq struct {
	ProductVariantID uuid.UUID   `json:"product_variant_id" validate:"required"`
	UOM              string      `json:"uom" validate:"required,oneof=pcs m2 m_lari box"`
	LengthCM         *float64    `json:"length_cm" validate:"omitempty,numeric,gt=0"`
	WidthCM          *float64    `json:"width_cm" validate:"omitempty,numeric,gt=0"`
	Quantity         int         `json:"quantity" validate:"required,gt=0"`
	DesignFileURL    *string     `json:"design_file_url" validate:"omitempty,url"`
	ProductionNotes  *string     `json:"production_notes" validate:"omitempty"`
	FinishingIDs     []uuid.UUID `json:"finishing_ids" validate:"omitempty"`
}

type CreateOrderReq struct {
	DesignerID    uuid.UUID      `json:"designer_id" validate:"required"`
	Notes         *string        `json:"notes" validate:"omitempty"`
	Items         []OrderItemReq `json:"items" validate:"required,dive"`
}

type PaymentProcessReq struct {
	ResellerID    *uuid.UUID `json:"reseller_id" validate:"omitempty"`
	CustomerName  string     `json:"customer_name" validate:"required"`
	CustomerPhone string     `json:"customer_phone" validate:"required"`
	PaymentType   string     `json:"payment_type" validate:"required,oneof=full dp"`
	AmountPaid    float64    `json:"amount_paid" validate:"required,numeric,gte=0"`
}

type CreateFinishingReq struct {
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"required,numeric,gte=0"`
}

type FinishingRes struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Price float64  `json:"price"`
}

type OrderItemRes struct {
	ID              uuid.UUID      `json:"id"`
	ProductVariantID uuid.UUID     `json:"product_variant_id"`
	ProductName      string         `json:"product_name"`
	VariantName      string         `json:"variant_name"`
	UOM              string         `json:"uom"`
	LengthCM         *float64       `json:"length_cm,omitempty"`
	WidthCM          *float64       `json:"width_cm,omitempty"`
	Quantity         int            `json:"quantity"`
	DesignFileURL    *string        `json:"design_file_url,omitempty"`
	ProductionNotes  *string        `json:"production_notes,omitempty"`
	PricePerUnit     float64        `json:"price_per_unit"`
	AdditionalCost   float64        `json:"additional_cost"`
	Subtotal         float64        `json:"subtotal"`
	Finishings       []FinishingRes `json:"finishings,omitempty"`
}

type OrderRes struct {
	ID                  uuid.UUID      `json:"id"`
	JobNumber            string         `json:"job_number"`
	InvoiceNumber        *string        `json:"invoice_number,omitempty"`
	ResellerID           *uuid.UUID     `json:"reseller_id,omitempty"`
	ResellerName         *string        `json:"reseller_name,omitempty"`
	DesignerID           uuid.UUID      `json:"designer_id"`
	DesignerName         string         `json:"designer_name"`
	CashierID            *uuid.UUID     `json:"cashier_id,omitempty"`
	CashierName          *string        `json:"cashier_name,omitempty"`
	CustomerName         *string        `json:"customer_name,omitempty"`
	CustomerPhone        *string        `json:"customer_phone,omitempty"`
	Status               string         `json:"status"`
	PaymentStatus        string         `json:"payment_status"`
	Notes                *string        `json:"notes,omitempty"`
	TotalAdditionalCost  float64        `json:"total_additional_cost"`
	TotalProductPrice    float64        `json:"total_product_price"`
	GrandTotal           float64        `json:"grand_total"`
	AmountPaid           float64        `json:"amount_paid"`
	OrderItems           []OrderItemRes `json:"order_items,omitempty"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}
