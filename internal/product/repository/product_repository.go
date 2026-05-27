package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Order("product_variants.is_default DESC, product_variants.created_at ASC")
		}).
		Preload("Variants.AttributeValues").
		Preload("Variants.AttributeValues.Attribute").
		Preload("Variants.PriceTiers").
		Preload("Variants.PriceTiers.CustomerLevel").
		First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindAll(ctx context.Context, params request.PaginationParam, search string, categoryID *uuid.UUID) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Product{})

	if categoryID != nil && *categoryID != uuid.Nil {
		db = db.Where("category_id = ?", *categoryID)
	}

	if search != "" {
		searchText := "%" + search + "%"
		db = db.Where("name ILIKE ? OR sku ILIKE ?", searchText, searchText)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("Category").
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Order("product_variants.is_default DESC, product_variants.created_at ASC")
		}).
		Preload("Variants.AttributeValues").
		Preload("Variants.AttributeValues.Attribute").
		Preload("Variants.PriceTiers").
		Preload("Variants.PriceTiers.CustomerLevel").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Order(params.GetSort()).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Product{}, "id = ?", id).Error
}

func (r *productRepository) CreateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	return r.db.WithContext(ctx).Create(variant).Error
}
