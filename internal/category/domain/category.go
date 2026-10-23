package domain

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Sequence      int            `gorm:"type:int;default:0" json:"sequence"`
	SPKCategoryID *uuid.UUID     `gorm:"type:uuid" json:"spk_category_id,omitempty"`
	SPKCategory   *Category      `gorm:"foreignKey:SPKCategoryID" json:"spk_category,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Category, int64, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	Reorder(ctx context.Context, items []request.ReorderItem) error
}

type CategoryUsecase interface {
	Create(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	FindAll(ctx context.Context, params request.PaginationParam, search string) ([]Category, int64, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	Reorder(ctx context.Context, items []request.ReorderItem) error
}
