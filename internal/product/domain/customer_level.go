package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerLevel struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name               string         `gorm:"type:varchar(100);not null" json:"name"`
	DiscountPercentage float64        `gorm:"type:decimal(5,2);not null;default:0" json:"discount_percentage"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (cl *CustomerLevel) BeforeCreate(tx *gorm.DB) error {
	if cl.ID == uuid.Nil {
		cl.ID = uuid.New()
	}
	return nil
}

type CustomerLevelRepository interface {
	Create(ctx context.Context, level *CustomerLevel) error
	FindByID(ctx context.Context, id uuid.UUID) (*CustomerLevel, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]CustomerLevel, int64, error)
	Update(ctx context.Context, level *CustomerLevel) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CustomerLevelUsecase interface {
	Create(ctx context.Context, level *CustomerLevel) error
	FindByID(ctx context.Context, id uuid.UUID) (*CustomerLevel, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]CustomerLevel, int64, error)
	Update(ctx context.Context, level *CustomerLevel) error
	Delete(ctx context.Context, id uuid.UUID) error
}
