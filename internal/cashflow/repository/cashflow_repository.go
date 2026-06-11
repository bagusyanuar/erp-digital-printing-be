package repository

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"gorm.io/gorm"
)

type cashFlowRepository struct {
	db *gorm.DB
}

func NewCashFlowRepository(db *gorm.DB) domain.CashFlowRepository {
	return &cashFlowRepository{db: db}
}

func (r *cashFlowRepository) Create(ctx context.Context, cashFlow *domain.CashFlow) error {
	return r.db.WithContext(ctx).Create(cashFlow).Error
}

func (r *cashFlowRepository) CreateTx(ctx context.Context, tx *gorm.DB, cashFlow *domain.CashFlow) error {
	if tx == nil {
		return r.Create(ctx, cashFlow)
	}
	return tx.WithContext(ctx).Create(cashFlow).Error
}

func (r *cashFlowRepository) FindAll(ctx context.Context, startDate time.Time, endDate time.Time) ([]domain.CashFlow, error) {
	var cashFlows []domain.CashFlow
	err := r.db.WithContext(ctx).
		Preload("Cashier").
		Where("transaction_date >= ? AND transaction_date <= ?", startDate, endDate).
		Order("transaction_date DESC, created_at DESC").
		Find(&cashFlows).Error
	if err != nil {
		return nil, err
	}
	return cashFlows, nil
}
