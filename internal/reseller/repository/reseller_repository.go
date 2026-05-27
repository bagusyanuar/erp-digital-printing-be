package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type resellerRepository struct {
	db *gorm.DB
}

func NewResellerRepository(db *gorm.DB) domain.ResellerRepository {
	return &resellerRepository{db: db}
}

func (r *resellerRepository) Create(ctx context.Context, reseller *domain.Reseller) error {
	return r.db.WithContext(ctx).Create(reseller).Error
}

func (r *resellerRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Reseller, error) {
	var reseller domain.Reseller
	if err := r.db.WithContext(ctx).Preload("CustomerLevel").First(&reseller, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &reseller, nil
}

func (r *resellerRepository) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Reseller, int64, error) {
	var resellers []domain.Reseller
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Reseller{})

	if search != "" {
		searchText := "%" + search + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR address ILIKE ?", 
			searchText, searchText, searchText, searchText)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("CustomerLevel").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Order(params.GetSort()).
		Find(&resellers).Error; err != nil {
		return nil, 0, err
	}

	return resellers, total, nil
}

func (r *resellerRepository) Update(ctx context.Context, reseller *domain.Reseller) error {
	return r.db.WithContext(ctx).Save(reseller).Error
}

func (r *resellerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Reseller{}, "id = ?", id).Error
}
