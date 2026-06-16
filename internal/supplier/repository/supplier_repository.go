package repository

import (
	"context"
	"errors"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) domain.SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

func (r *supplierRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier
	if err := r.db.WithContext(ctx).First(&supplier, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *supplierRepository) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Supplier, int64, error) {
	var suppliers []domain.Supplier
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Supplier{})

	if search != "" {
		searchText := "%" + search + "%"
		db = db.Where("name ILIKE ? OR contact_name ILIKE ? OR email ILIKE ?", searchText, searchText, searchText)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Order(params.GetSort()).
		Find(&suppliers).Error; err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Supplier{}, "id = ?", id).Error
}

func (r *supplierRepository) FindByName(ctx context.Context, name string) (*domain.Supplier, error) {
	var supplier domain.Supplier
	if err := r.db.WithContext(ctx).First(&supplier, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &supplier, nil
}
