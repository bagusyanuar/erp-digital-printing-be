package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	cfDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type expenseUsecase struct {
	repo   domain.ExpenseRepository
	cfRepo cfDomain.CashFlowRepository
	db     *gorm.DB
	logger *zap.Logger
}

func NewExpenseUsecase(repo domain.ExpenseRepository, cfRepo cfDomain.CashFlowRepository, db *gorm.DB, logger *zap.Logger) domain.ExpenseUsecase {
	return &expenseUsecase{
		repo:   repo,
		cfRepo: cfRepo,
		db:     db,
		logger: logger,
	}
}

func (u *expenseUsecase) CreateCategory(ctx context.Context, category *domain.ExpenseCategory) error {
	if category.Group != domain.GroupProduction && category.Group != domain.GroupOperational {
		return errors.New("invalid category group")
	}

	if category.Group == domain.GroupProduction && category.ProductCategoryID == nil {
		return errors.New("product category ID is required for production group")
	}

	if category.Group == domain.GroupOperational {
		category.ProductCategoryID = nil
	}

	return u.repo.CreateCategory(ctx, category)
}

func (u *expenseUsecase) FindCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ExpenseCategory, error) {
	return u.repo.FindCategoryByID(ctx, id)
}

func (u *expenseUsecase) FindAllCategories(ctx context.Context, group string) ([]domain.ExpenseCategory, error) {
	return u.repo.FindAllCategories(ctx, group)
}

func (u *expenseUsecase) UpdateCategory(ctx context.Context, category *domain.ExpenseCategory) error {
	if category.Group != domain.GroupProduction && category.Group != domain.GroupOperational {
		return errors.New("invalid category group")
	}

	if category.Group == domain.GroupProduction && category.ProductCategoryID == nil {
		return errors.New("product category ID is required for production group")
	}

	if category.Group == domain.GroupOperational {
		category.ProductCategoryID = nil
	}

	return u.repo.UpdateCategory(ctx, category)
}

func (u *expenseUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	hasExpenses, err := u.repo.HasAssociatedExpenses(ctx, id)
	if err != nil {
		return err
	}
	if hasExpenses {
		return errors.New("cannot delete category with associated expenses")
	}
	return u.repo.DeleteCategory(ctx, id)
}

func (u *expenseUsecase) CreateExpense(ctx context.Context, expense *domain.Expense) error {
	category, err := u.repo.FindCategoryByID(ctx, expense.ExpenseCategoryID)
	if err != nil {
		return fmt.Errorf("expense category not found: %w", err)
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := u.cfRepo.FindAccountByNameWithLock(ctx, tx, expense.PaymentMethod)
		if err != nil {
			return fmt.Errorf("failed to lock cash account: %w", err)
		}

		acc.Balance -= expense.Amount
		if err := u.cfRepo.UpdateAccount(ctx, tx, acc); err != nil {
			return fmt.Errorf("failed to update cash account balance: %w", err)
		}

		if err := u.repo.CreateExpenseTx(ctx, tx, expense); err != nil {
			return fmt.Errorf("failed to create expense record: %w", err)
		}

		desc := fmt.Sprintf("Pengeluaran: %s", category.Name)
		if expense.Description != nil && *expense.Description != "" {
			desc = fmt.Sprintf("Pengeluaran: %s - %s", category.Name, *expense.Description)
		}

		cf := &cfDomain.CashFlow{
			ID:              uuid.New(),
			TransactionDate: expense.ExpenseDate,
			ReferenceType:   "EXPENSE",
			ReferenceID:     &expense.ID,
			Type:            cfDomain.TypeCredit,
			Amount:          expense.Amount,
			PaymentMethod:   expense.PaymentMethod,
			Description:     &desc,
			CashierID:       expense.CashierID,
		}

		if err := u.cfRepo.CreateTx(ctx, tx, cf); err != nil {
			return fmt.Errorf("failed to create cash flow record: %w", err)
		}

		return nil
	})
}

func (u *expenseUsecase) FindAllExpenses(ctx context.Context, filter domain.ExpenseFilter) ([]domain.Expense, int64, error) {
	return u.repo.FindAllExpenses(ctx, filter)
}

func (u *expenseUsecase) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	expense, err := u.repo.FindExpenseByID(ctx, id)
	if err != nil {
		return fmt.Errorf("expense record not found: %w", err)
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := u.cfRepo.FindAccountByNameWithLock(ctx, tx, expense.PaymentMethod)
		if err != nil {
			return fmt.Errorf("failed to lock cash account: %w", err)
		}

		acc.Balance += expense.Amount
		if err := u.cfRepo.UpdateAccount(ctx, tx, acc); err != nil {
			return fmt.Errorf("failed to restore cash account balance: %w", err)
		}

		// Delete associated cash flow record (soft-delete is handled automatically if model is deleted with gorm)
		if err := tx.Where("reference_type = ? AND reference_id = ?", "EXPENSE", expense.ID).Delete(&cfDomain.CashFlow{}).Error; err != nil {
			return fmt.Errorf("failed to delete cash flow record: %w", err)
		}

		if err := u.repo.DeleteExpenseTx(ctx, tx, expense.ID); err != nil {
			return fmt.Errorf("failed to delete expense record: %w", err)
		}

		return nil
	})
}

func (u *expenseUsecase) GetSummary(ctx context.Context, startDate *time.Time, endDate *time.Time) (*domain.ExpenseSummaryRes, error) {
	return u.repo.GetSummary(ctx, startDate, endDate)
}

func (u *expenseUsecase) GetByProductCategory(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]domain.ExpenseByProductCategoryRes, error) {
	return u.repo.GetByProductCategory(ctx, startDate, endDate)
}
