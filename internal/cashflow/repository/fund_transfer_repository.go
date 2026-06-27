package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fundTransferRepository struct {
	db *gorm.DB
}

func NewFundTransferRepository(db *gorm.DB) domain.FundTransferRepository {
	return &fundTransferRepository{db: db}
}

func (r *fundTransferRepository) CreateTx(ctx context.Context, tx *gorm.DB, transfer *domain.FundTransfer) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(transfer).Error
}

func (r *fundTransferRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.FundTransfer, error) {
	var transfer domain.FundTransfer
	err := r.db.WithContext(ctx).
		Preload("FromAccount").
		Preload("ToAccount").
		Preload("Cashier").
		First(&transfer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *fundTransferRepository) FindAll(ctx context.Context, filter domain.FundTransferFilter) ([]domain.FundTransfer, int64, error) {
	var transfers []domain.FundTransfer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.FundTransfer{})

	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() {
		query = query.Where("transfer_date >= ? AND transfer_date <= ?", filter.StartDate, filter.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Preload("FromAccount").
		Preload("ToAccount").
		Preload("Cashier").
		Order("transfer_date DESC, created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&transfers).Error

	if err != nil {
		return nil, 0, err
	}

	return transfers, total, nil
}

func (r *fundTransferRepository) DeleteTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Delete(&domain.FundTransfer{}, "id = ?", id).Error
}

func (r *fundTransferRepository) GetWidgetsData(ctx context.Context, startDate, endDate time.Time) (float64, map[uuid.UUID]float64, error) {
	var total float64

	queryTotal := r.db.WithContext(ctx).Model(&domain.FundTransfer{})
	if !startDate.IsZero() && !endDate.IsZero() {
		queryTotal = queryTotal.Where("transfer_date >= ? AND transfer_date <= ?", startDate, endDate)
	}

	err := queryTotal.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	if err != nil {
		return 0, nil, err
	}

	type Row struct {
		ToAccountID uuid.UUID
		SumAmount   float64
	}
	var rows []Row

	queryBreakdown := r.db.WithContext(ctx).Model(&domain.FundTransfer{})
	if !startDate.IsZero() && !endDate.IsZero() {
		queryBreakdown = queryBreakdown.Where("transfer_date >= ? AND transfer_date <= ?", startDate, endDate)
	}

	err = queryBreakdown.
		Select("to_account_id, COALESCE(SUM(amount), 0) as sum_amount").
		Group("to_account_id").
		Scan(&rows).Error
	if err != nil {
		return 0, nil, err
	}

	breakdown := make(map[uuid.UUID]float64)
	for _, row := range rows {
		breakdown[row.ToAccountID] = row.SumAmount
	}

	return total, breakdown, nil
}
