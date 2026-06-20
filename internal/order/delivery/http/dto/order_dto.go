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
	DesignerID    uuid.UUID      `json:"designer_id,omitempty" validate:"omitempty"`
	ResellerID    *uuid.UUID     `json:"reseller_id" validate:"omitempty"`
	CustomerName  *string        `json:"customer_name" validate:"omitempty"`
	CustomerPhone *string        `json:"customer_phone" validate:"omitempty"`
	Notes         *string        `json:"notes" validate:"omitempty"`
	Items         []OrderItemReq `json:"items" validate:"required,dive"`
}

type PaymentItemReq struct {
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash qris transfer tempo"`
	AmountPaid    float64 `json:"amount_paid" validate:"required,numeric,gt=0"`
}

type PaymentProcessReq struct {
	ResellerID    *uuid.UUID       `json:"reseller_id" validate:"omitempty"`
	CustomerName  string           `json:"customer_name" validate:"required"`
	CustomerPhone string           `json:"customer_phone" validate:"required"`
	Payments      []PaymentItemReq `json:"payments" validate:"required,min=1,dive"`
}

type OrderRepayReq struct {
	Payments []PaymentItemReq `json:"payments" validate:"required,min=1,dive"`
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

type OrderPaymentRes struct {
	ID            uuid.UUID `json:"id"`
	CashierID     uuid.UUID `json:"cashier_id"`
	CashierName   string    `json:"cashier_name"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentType   string    `json:"payment_type"`
	PaymentNumber int       `json:"payment_number"`
	CreatedAt     string    `json:"created_at"`
}

type ResellerRes struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	Email       *string   `json:"email,omitempty"`
	Address     string    `json:"address"`
	CreditLimit float64   `json:"credit_limit"`
}

type OrderRes struct {
	ID                  uuid.UUID         `json:"id"`
	JobNumber           string            `json:"job_number"`
	InvoiceNumber       *string           `json:"invoice_number"`
	ResellerID          *uuid.UUID        `json:"reseller_id"`
	ResellerName        *string           `json:"reseller_name"`
	Reseller            *ResellerRes      `json:"reseller"`
	DesignerID          uuid.UUID         `json:"designer_id"`
	DesignerName        string            `json:"designer_name"`
	CashierID           *uuid.UUID        `json:"cashier_id"`
	CashierName         *string           `json:"cashier_name"`
	CustomerName        *string           `json:"customer_name"`
	CustomerPhone       *string           `json:"customer_phone"`
	Status              string            `json:"status"`
	PaymentStatus       string            `json:"payment_status"`
	Notes               *string           `json:"notes,omitempty"`
	TotalAdditionalCost float64           `json:"total_additional_cost"`
	TotalProductPrice   float64           `json:"total_product_price"`
	GrandTotal          float64           `json:"grand_total"`
	AmountPaid          float64           `json:"amount_paid"`
	OrderItems          []OrderItemRes    `json:"order_items,omitempty"`
	OrderPayments       []OrderPaymentRes `json:"order_payments,omitempty"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
}

type UpdateOrderStatusReq struct {
	Status string `json:"status" validate:"required,oneof=DRAFT PENDING_PAYMENT IN_PRODUCTION READY_FOR_PICKUP COMPLETED CANCELLED"`
}

type OrderReportsStatusNotaRes struct {
	Lunas      int64 `json:"lunas"`
	BelumLunas int64 `json:"belum_lunas"`
}

type OrderReportsWidgetsRes struct {
	OmsetPenjualan     float64                   `json:"omset_penjualan"`
	VolumeTransaksi    int64                     `json:"volume_transaksi"`
	TotalProdukTerjual int64                     `json:"total_produk_terjual"`
	StatusNota         OrderReportsStatusNotaRes `json:"status_nota"`
}

