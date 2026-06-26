package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/capital/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type capitalRepository struct {
	db *gorm.DB
}

func NewCapitalRepository(db *gorm.DB) domain.CapitalRepository {
	return &capitalRepository{db: db}
}

func (r *capitalRepository) CreateTx(ctx context.Context, tx *gorm.DB, transaction *domain.CapitalTransaction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(transaction).Error
}

func (r *capitalRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.CapitalTransaction, error) {
	return r.FindByIDTx(ctx, nil, id)
}

func (r *capitalRepository) FindByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*domain.CapitalTransaction, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var transaction domain.CapitalTransaction
	if err := db.WithContext(ctx).First(&transaction, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *capitalRepository) FindAll(ctx context.Context, filter domain.CapitalFilter) ([]domain.CapitalTransaction, int64, error) {
	var transactions []domain.CapitalTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.CapitalTransaction{})

	if filter.StartDate != nil {
		query = query.Where("transaction_date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("transaction_date <= ?", *filter.EndDate)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("description ILIKE ? OR payment_method ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Preload("Creator").
		Order("transaction_date DESC, created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *capitalRepository) DeleteTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Delete(&domain.CapitalTransaction{}, "id = ?", id).Error
}
