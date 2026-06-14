package repository

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) domain.ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *expenseRepository) CreateCategory(ctx context.Context, category *domain.ExpenseCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *expenseRepository) FindCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ExpenseCategory, error) {
	var category domain.ExpenseCategory
	if err := r.db.WithContext(ctx).
		Preload("ProductCategory").
		First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *expenseRepository) FindAllCategories(ctx context.Context, group string) ([]domain.ExpenseCategory, error) {
	var categories []domain.ExpenseCategory
	db := r.db.WithContext(ctx).Preload("ProductCategory")
	if group != "" {
		db = db.Where(`"group" = ?`, group)
	}
	if err := db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *expenseRepository) UpdateCategory(ctx context.Context, category *domain.ExpenseCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *expenseRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ExpenseCategory{}, "id = ?", id).Error
}

func (r *expenseRepository) HasAssociatedExpenses(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Expense{}).Where("expense_category_id = ?", categoryID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *expenseRepository) CreateExpenseTx(ctx context.Context, tx *gorm.DB, expense *domain.Expense) error {
	return tx.WithContext(ctx).Create(expense).Error
}

func (r *expenseRepository) FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.Expense, error) {
	var expense domain.Expense
	if err := r.db.WithContext(ctx).
		Preload("ExpenseCategory.ProductCategory").
		First(&expense, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) FindExpenseByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*domain.Expense, error) {
	var expense domain.Expense
	if err := tx.WithContext(ctx).
		Preload("ExpenseCategory.ProductCategory").
		First(&expense, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) FindAllExpenses(ctx context.Context, filter domain.ExpenseFilter) ([]domain.Expense, int64, error) {
	var expenses []domain.Expense
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Expense{}).
		Joins("JOIN expense_categories ON expenses.expense_category_id = expense_categories.id").
		Preload("ExpenseCategory.ProductCategory")

	if filter.Group != "" {
		db = db.Where("expense_categories.group = ?", filter.Group)
	}
	if filter.CategoryID != nil {
		db = db.Where("expenses.expense_category_id = ?", *filter.CategoryID)
	}
	if filter.StartDate != nil {
		db = db.Where("expenses.expense_date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		db = db.Where("expenses.expense_date <= ?", *filter.EndDate)
	}
	if filter.Search != "" {
		db = db.Where("expenses.description ILIKE ?", "%"+filter.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * filter.Limit
	}
	limit := 10
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	if err := db.Order("expenses.expense_date desc").
		Limit(limit).
		Offset(offset).
		Find(&expenses).Error; err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

func (r *expenseRepository) DeleteExpenseTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	return tx.WithContext(ctx).Delete(&domain.Expense{}, "id = ?", id).Error
}

func (r *expenseRepository) GetSummary(ctx context.Context, startDate *time.Time, endDate *time.Time) (*domain.ExpenseSummaryRes, error) {
	var summary domain.ExpenseSummaryRes

	queryProd := r.db.WithContext(ctx).Model(&domain.Expense{}).
		Joins("JOIN expense_categories ON expenses.expense_category_id = expense_categories.id").
		Where("expense_categories.group = ?", domain.GroupProduction)

	queryOps := r.db.WithContext(ctx).Model(&domain.Expense{}).
		Joins("JOIN expense_categories ON expenses.expense_category_id = expense_categories.id").
		Where("expense_categories.group = ?", domain.GroupOperational)

	if startDate != nil {
		queryProd = queryProd.Where("expenses.expense_date >= ?", *startDate)
		queryOps = queryOps.Where("expenses.expense_date >= ?", *startDate)
	}
	if endDate != nil {
		queryProd = queryProd.Where("expenses.expense_date <= ?", *endDate)
		queryOps = queryOps.Where("expenses.expense_date <= ?", *endDate)
	}

	var totalProd, totalOps float64
	queryProd.Select("COALESCE(SUM(expenses.amount), 0)").Row().Scan(&totalProd)
	queryOps.Select("COALESCE(SUM(expenses.amount), 0)").Row().Scan(&totalOps)

	summary.TotalProduction = totalProd
	summary.TotalOperational = totalOps
	summary.TotalExpense = totalProd + totalOps

	return &summary, nil
}

func (r *expenseRepository) GetByProductCategory(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]domain.ExpenseByProductCategoryRes, error) {
	var results []domain.ExpenseByProductCategoryRes

	db := r.db.WithContext(ctx).Model(&domain.Expense{}).
		Select("expense_categories.product_category_id, categories.name as product_category_name, COALESCE(SUM(expenses.amount), 0) as total_amount").
		Joins("JOIN expense_categories ON expenses.expense_category_id = expense_categories.id").
		Joins("JOIN categories ON expense_categories.product_category_id = categories.id").
		Where("expense_categories.group = ?", domain.GroupProduction).
		Group("expense_categories.product_category_id, categories.name")

	if startDate != nil {
		db = db.Where("expenses.expense_date >= ?", *startDate)
	}
	if endDate != nil {
		db = db.Where("expenses.expense_date <= ?", *endDate)
	}

	if err := db.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
