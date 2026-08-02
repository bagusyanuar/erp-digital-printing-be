package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/category/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Category{}).Where("id != ?", "00000000-0000-0000-0000-000000000000")

	if search != "" {
		searchText := "%" + search + "%"
		db = db.Where("name ILIKE ?", searchText)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Order("sequence ASC").
		Order(params.GetSort()).
		Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id).Error
}

func (r *categoryRepository) Reorder(ctx context.Context, items []request.ReorderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update sequence as requested
		for _, item := range items {
			if err := tx.Model(&domain.Category{}).Where("id = ?", item.ID).Update("sequence", item.Sequence).Error; err != nil {
				return err
			}
		}

		// Brute-force Auto Resync
		var categories []domain.Category
		if err := tx.Order("sequence ASC, created_at DESC").Find(&categories).Error; err != nil {
			return err
		}

		for i, cat := range categories {
			if err := tx.Model(&domain.Category{}).Where("id = ?", cat.ID).Update("sequence", i+1).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
