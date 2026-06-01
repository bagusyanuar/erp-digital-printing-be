package dto

import "github.com/google/uuid"

type SPKItemRes struct {
	ID              uuid.UUID      `json:"id"`
	ProductName      string         `json:"product_name"`
	VariantName      string         `json:"variant_name"`
	UOM              string         `json:"uom"`
	LengthCM         *float64       `json:"length_cm"`
	WidthCM          *float64       `json:"width_cm"`
	Quantity         int            `json:"quantity"`
	DesignFileURL    *string        `json:"design_file_url,omitempty"`
	ProductionNotes  *string        `json:"production_notes,omitempty"`
	Finishings       []FinishingRes `json:"finishings,omitempty"`
}

type SPKByCategoryRes struct {
	CategoryID   uuid.UUID    `json:"category_id"`
	CategoryName string       `json:"category_name"`
	Items        []SPKItemRes `json:"items"`
}

type OrderSPKRes struct {
	OrderID        uuid.UUID          `json:"order_id"`
	JobNumber      string             `json:"job_number"`
	InvoiceNumber  *string            `json:"invoice_number,omitempty"`
	CustomerName   *string            `json:"customer_name,omitempty"`
	CustomerPhone  *string            `json:"customer_phone,omitempty"`
	Status         string             `json:"status"`
	SPKByCategory  []SPKByCategoryRes `json:"spk_by_category"`
}
