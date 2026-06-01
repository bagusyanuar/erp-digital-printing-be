package domain

import (
	"context"
	"time"

	productDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	resellerDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/domain"
	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order Status Constants
const (
	StatusDraft          = "DRAFT"
	StatusPendingPayment = "PENDING_PAYMENT"
	StatusInProduction   = "IN_PRODUCTION"
	StatusReadyForPickup = "READY_FOR_PICKUP"
	StatusCompleted      = "COMPLETED"
	StatusCancelled      = "CANCELLED"
)

// Payment Status Constants
const (
	PaymentStatusUnpaid      = "UNPAID"
	PaymentStatusDownPayment = "DOWN_PAYMENT"
	PaymentStatusPaid        = "PAID"
)

// Finishing model
type Finishing struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Price     float64        `gorm:"type:decimal(15,2);not null;default:0" json:"price"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (f *Finishing) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// Order model
type Order struct {
	ID                   uuid.UUID                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	JobNumber            string                   `gorm:"type:varchar(100);unique;not null" json:"job_number"`
	InvoiceNumber        *string                  `gorm:"type:varchar(100);unique" json:"invoice_number,omitempty"`
	ResellerID           *uuid.UUID               `gorm:"type:uuid" json:"reseller_id,omitempty"`
	Reseller             *resellerDomain.Reseller `gorm:"foreignKey:ResellerID" json:"reseller,omitempty"`
	DesignerID           uuid.UUID                `gorm:"type:uuid;not null" json:"designer_id"`
	Designer             *userDomain.User         `gorm:"foreignKey:DesignerID" json:"designer,omitempty"`
	CashierID            *uuid.UUID               `gorm:"type:uuid" json:"cashier_id,omitempty"`
	Cashier              *userDomain.User         `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	CustomerName         *string                  `gorm:"type:varchar(255)" json:"customer_name,omitempty"`
	CustomerPhone        *string                  `gorm:"type:varchar(20)" json:"customer_phone,omitempty"`
	Status               string                   `gorm:"type:varchar(50);not null;default:'DRAFT'" json:"status"`
	PaymentStatus        string                   `gorm:"type:varchar(50);not null;default:'UNPAID'" json:"payment_status"`
	Notes                *string                  `gorm:"type:text" json:"notes,omitempty"`
	TotalAdditionalCost  float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"total_additional_cost"`
	TotalProductPrice    float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"total_product_price"`
	GrandTotal           float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"grand_total"`
	AmountPaid           float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"amount_paid"`
	OrderItems           []OrderItem              `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"order_items,omitempty"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
	DeletedAt            gorm.DeletedAt           `gorm:"index" json:"deleted_at,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// OrderItem model
type OrderItem struct {
	ID               uuid.UUID                     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID          uuid.UUID                     `gorm:"type:uuid;not null" json:"order_id"`
	ProductVariantID uuid.UUID                     `gorm:"type:uuid;not null" json:"product_variant_id"`
	ProductVariant   *productDomain.ProductVariant `gorm:"foreignKey:ProductVariantID" json:"product_variant,omitempty"`
	UOM              string                        `gorm:"type:varchar(50);not null" json:"uom"`
	LengthCM         *float64                      `gorm:"type:decimal(10,2)" json:"length_cm,omitempty"`
	WidthCM          *float64                      `gorm:"type:decimal(10,2)" json:"width_cm,omitempty"`
	Quantity         int                           `gorm:"type:int;not null" json:"quantity"`
	DesignFileURL    *string                       `gorm:"type:text" json:"design_file_url,omitempty"`
	ProductionNotes  *string                       `gorm:"type:text" json:"production_notes,omitempty"`
	PricePerUnit     float64                       `gorm:"type:decimal(15,2);not null;default:0" json:"price_per_unit"`
	AdditionalCost   float64                       `gorm:"type:decimal(15,2);not null;default:0" json:"additional_cost"`
	Subtotal         float64                       `gorm:"type:decimal(15,2);not null;default:0" json:"subtotal"`
	Finishings       []Finishing                   `gorm:"many2many:order_item_finishings;" json:"finishings,omitempty"`
	CreatedAt        time.Time                     `json:"created_at"`
	UpdatedAt        time.Time                     `json:"updated_at"`
	DeletedAt        gorm.DeletedAt                `gorm:"index" json:"deleted_at,omitempty"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}

// OrderRepository interface
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindAll(ctx context.Context, params request.PaginationParam, statuses []string, designerID *uuid.UUID) ([]Order, int64, error)
	GetNextJobSeq(ctx context.Context, dateStr string) (int, error)
	GetNextInvSeq(ctx context.Context, dateStr string) (int, error)
	FindFinishingsByIDs(ctx context.Context, ids []uuid.UUID) ([]Finishing, error)
	CreateFinishing(ctx context.Context, finishing *Finishing) error
	FindAllFinishings(ctx context.Context) ([]Finishing, error)
}

// OrderUsecase interface
type OrderUsecase interface {
	SaveDraft(ctx context.Context, order *Order) error
	SubmitToCashier(ctx context.Context, order *Order) error
	SubmitExistingToCashier(ctx context.Context, orderID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindAll(ctx context.Context, params request.PaginationParam, statuses []string, designerID *uuid.UUID) ([]Order, int64, error)
	ProcessPayment(ctx context.Context, orderID uuid.UUID, cashierID uuid.UUID, resellerID *uuid.UUID, customerName string, customerPhone string, paymentType string, amountPaid float64) (*Order, error)
	CreateFinishing(ctx context.Context, finishing *Finishing) error
	FindAllFinishings(ctx context.Context) ([]Finishing, error)
}
