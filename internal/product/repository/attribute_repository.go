package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type attributeRepository struct {
	db *gorm.DB
}

func NewAttributeRepository(db *gorm.DB) domain.AttributeRepository {
	return &attributeRepository{db: db}
}

func (r *attributeRepository) Create(ctx context.Context, attribute *domain.Attribute) error {
	return r.db.WithContext(ctx).Create(attribute).Error
}

func (r *attributeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Attribute, error) {
	var attribute domain.Attribute
	if err := r.db.WithContext(ctx).Preload("Options").First(&attribute, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &attribute, nil
}

func (r *attributeRepository) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Attribute, int64, error) {
	var attributes []domain.Attribute
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Attribute{})

	if search != "" {
		searchText := "%" + search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", searchText, searchText)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Options").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Order(params.GetSort()).
		Find(&attributes).Error; err != nil {
		return nil, 0, err
	}

	return attributes, total, nil
}

func (r *attributeRepository) Update(ctx context.Context, attribute *domain.Attribute) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(attribute).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Always sync options relationship. If empty, GORM will clear the relation.
	if err := tx.Model(attribute).Association("Options").Replace(attribute.Options); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *attributeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Attribute{}, "id = ?", id).Error
}
