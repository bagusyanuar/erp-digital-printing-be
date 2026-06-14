package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *cashFlowRepository) FindAll(ctx context.Context, filter domain.CashFlowFilter) ([]domain.CashFlow, int64, error) {
	var cashFlows []domain.CashFlow
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.CashFlow{})

	// Apply Filters
	query = query.Where("transaction_date >= ? AND transaction_date <= ?", filter.StartDate, filter.EndDate)

	if filter.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filter.PaymentMethod)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.ReferenceType != "" {
		query = query.Where("reference_type = ?", filter.ReferenceType)
	}

	if filter.CashierID != nil {
		query = query.Where("cashier_id = ?", *filter.CashierID)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("description ILIKE ? OR customer_name ILIKE ? OR invoice_number ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Count total rows matching filters (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply Pagination
	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Preload("Cashier").
		Order("transaction_date DESC, created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&cashFlows).Error

	if err != nil {
		return nil, 0, err
	}

	return cashFlows, total, nil
}

func (r *cashFlowRepository) GetSummary(ctx context.Context, filter domain.CashFlowFilter) (*domain.CashFlowSummaryRes, error) {
	// Buat query dasar dengan filter yang sama persis
	query := r.db.WithContext(ctx).Model(&domain.CashFlow{})

	query = query.Where("transaction_date >= ? AND transaction_date <= ?", filter.StartDate, filter.EndDate)

	if filter.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filter.PaymentMethod)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.ReferenceType != "" {
		query = query.Where("reference_type = ?", filter.ReferenceType)
	}

	if filter.CashierID != nil {
		query = query.Where("cashier_id = ?", *filter.CashierID)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("description ILIKE ? OR customer_name ILIKE ? OR invoice_number ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Agregasikan berdasarkan payment_method dan type menggunakan query group/struct biasa
	type AggResult struct {
		PaymentMethod string
		Type          string
		TotalAmount   float64
	}

	var results []AggResult
	err := query.
		Select("payment_method, type, SUM(amount) as total_amount").
		Group("payment_method, type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Buat response data
	var totalDebit float64
	var totalCredit float64
	detailsByMethod := make(map[string]domain.CashFlowMethodDetail)

	for _, res := range results {
		method := res.PaymentMethod
		detail, exists := detailsByMethod[method]
		if !exists {
			detail = domain.CashFlowMethodDetail{}
		}

		if res.Type == domain.TypeDebit {
			detail.Debit += res.TotalAmount
			totalDebit += res.TotalAmount
		} else if res.Type == domain.TypeCredit {
			detail.Credit += res.TotalAmount
			totalCredit += res.TotalAmount
		}
		detail.Balance = detail.Debit - detail.Credit
		detailsByMethod[method] = detail
	}

	return &domain.CashFlowSummaryRes{
		Summary: domain.CashFlowSummary{
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
			NetBalance:  totalDebit - totalCredit,
		},
		DetailsByMethod: detailsByMethod,
	}, nil
}

func (r *cashFlowRepository) FindAllAccounts(ctx context.Context) ([]domain.CashAccount, error) {
	var accounts []domain.CashAccount
	err := r.db.WithContext(ctx).Order("name ASC").Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *cashFlowRepository) FindAccountByNameWithLock(ctx context.Context, tx *gorm.DB, name string) (*domain.CashAccount, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var account domain.CashAccount
	err := db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&account, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *cashFlowRepository) UpdateAccount(ctx context.Context, tx *gorm.DB, account *domain.CashAccount) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Save(account).Error
}
