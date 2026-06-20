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
	if len(expense.Items) == 0 {
		return errors.New("expense must have at least one item")
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 0. Auto-append discount item if present
		if expense.Discount > 0 {
			discountCategoryIdx := uuid.MustParse("00000000-0000-0000-0000-000000000000")
			desc := "Potongan Pembelian (Diskon)"
			expense.Items = append(expense.Items, domain.ExpenseItem{
				ID:                uuid.New(),
				ExpenseCategoryID: discountCategoryIdx,
				Description:       &desc,
				Amount:            -expense.Discount,
			})
		}

		// 1. Calculate amount & validate items
		var totalBelanja float64
		for i := range expense.Items {
			item := &expense.Items[i]
			if item.ExpenseCategoryID == uuid.Nil {
				return errors.New("each expense item must have a category ID")
			}
			
			// Verify category exists
			_, err := u.repo.FindCategoryByID(ctx, item.ExpenseCategoryID)
			if err != nil {
				return fmt.Errorf("expense category not found for item %d: %w", i, err)
			}

			if item.ID == uuid.Nil {
				item.ID = uuid.New()
			}
			totalBelanja += item.Amount
		}
		expense.Amount = totalBelanja

		// 2. Calculate initial payments
		var totalBayar float64
		for _, p := range expense.Payments {
			if p.Amount < 0 {
				return errors.New("payment amount cannot be negative")
			}
			totalBayar += p.Amount
		}

		if totalBayar > totalBelanja {
			return fmt.Errorf("initial payment amount (%.2f) exceeds total bill (%.2f)", totalBayar, totalBelanja)
		}

		// 3. Determine Status
		if totalBayar == totalBelanja {
			expense.Status = domain.StatusPaid
		} else if totalBayar > 0 {
			expense.Status = domain.StatusPartial
		} else {
			expense.Status = domain.StatusUnpaid
		}

		// 4. Generate Expense Number
		dateStr := time.Now().Format("20060102")
		var count int64
		if err := tx.Model(&domain.Expense{}).
			Where("expense_number LIKE ?", "EXP/"+dateStr+"/%").
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to generate expense number: %w", err)
		}
		expense.ExpenseNumber = fmt.Sprintf("EXP/%s/%04d", dateStr, count+1)

		if expense.ID == uuid.Nil {
			expense.ID = uuid.New()
		}
		if expense.ExpenseDate.IsZero() {
			expense.ExpenseDate = time.Now()
		}

		// 5. Save header
		if err := tx.Omit("Items", "Payments").Create(expense).Error; err != nil {
			return fmt.Errorf("failed to create expense: %w", err)
		}

		// 6. Save items
		for i := range expense.Items {
			expense.Items[i].ExpenseID = expense.ID
		}
		if len(expense.Items) > 0 {
			if err := tx.Create(&expense.Items).Error; err != nil {
				return fmt.Errorf("failed to create expense items: %w", err)
			}
		}

		// 7. Process payments
		for i := range expense.Payments {
			payment := &expense.Payments[i]
			if payment.Amount <= 0 {
				continue
			}
			payment.ID = uuid.New()
			payment.ExpenseID = expense.ID
			payment.CashierID = expense.CashierID
			if payment.PaymentDate.IsZero() {
				payment.PaymentDate = time.Now()
			}

			// Create payment log
			if err := tx.Create(payment).Error; err != nil {
				return fmt.Errorf("failed to save payment: %w", err)
			}

			// Lock Cash Account and deduct balance
			acc, err := u.cfRepo.FindAccountByNameWithLock(ctx, tx, payment.PaymentMethod)
			if err != nil {
				return fmt.Errorf("cash account %s not found: %w", payment.PaymentMethod, err)
			}
			acc.Balance -= payment.Amount
			if err := u.cfRepo.UpdateAccount(ctx, tx, acc); err != nil {
				return fmt.Errorf("failed to update cash account balance: %w", err)
			}

			// Create cash flow entry
			desc := fmt.Sprintf("Pengeluaran Nota %s (%s)", expense.ExpenseNumber, payment.PaymentMethod)
			if expense.InvoiceNumber != nil && *expense.InvoiceNumber != "" {
				desc = fmt.Sprintf("Pengeluaran Nota %s - Inv %s (%s)", expense.ExpenseNumber, *expense.InvoiceNumber, payment.PaymentMethod)
			}
			cf := &cfDomain.CashFlow{
				ID:              uuid.New(),
				TransactionDate: payment.PaymentDate,
				ReferenceType:   "EXPENSE_PAYMENT",
				ReferenceID:     &payment.ID,
				Type:            cfDomain.TypeCredit,
				Amount:          payment.Amount,
				PaymentMethod:   payment.PaymentMethod,
				Description:     &desc,
				CashierID:       payment.CashierID,
			}
			if err := u.cfRepo.CreateTx(ctx, tx, cf); err != nil {
				return fmt.Errorf("failed to record cash flow: %w", err)
			}
		}

		return nil
	})
}

func (u *expenseUsecase) PayInstallment(ctx context.Context, expenseID uuid.UUID, cashierID uuid.UUID, payments []domain.ExpensePayment) error {
	if len(payments) == 0 {
		return errors.New("at least one payment is required")
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		expense, err := u.repo.FindExpenseByIDTx(ctx, tx, expenseID)
		if err != nil {
			return fmt.Errorf("expense not found: %w", err)
		}

		if expense.Status == domain.StatusPaid {
			return errors.New("expense is already fully paid")
		}

		var totalPaidSoFar float64
		for _, p := range expense.Payments {
			totalPaidSoFar += p.Amount
		}

		var totalNewPayment float64
		for _, p := range payments {
			if p.Amount <= 0 {
				return errors.New("payment amount must be greater than 0")
			}
			totalNewPayment += p.Amount
		}

		remainingBill := expense.Amount - totalPaidSoFar
		if totalNewPayment > remainingBill {
			return fmt.Errorf("payment amount (%.2f) exceeds remaining bill (%.2f)", totalNewPayment, remainingBill)
		}

		// Save each payment, update cash account and register cash flows
		for i := range payments {
			p := &payments[i]
			p.ID = uuid.New()
			p.ExpenseID = expense.ID
			p.CashierID = cashierID
			if p.PaymentDate.IsZero() {
				p.PaymentDate = time.Now()
			}

			if err := tx.Create(p).Error; err != nil {
				return fmt.Errorf("failed to create payment record: %w", err)
			}

			// Lock Cash Account and deduct balance
			acc, err := u.cfRepo.FindAccountByNameWithLock(ctx, tx, p.PaymentMethod)
			if err != nil {
				return fmt.Errorf("cash account %s not found: %w", p.PaymentMethod, err)
			}
			acc.Balance -= p.Amount
			if err := u.cfRepo.UpdateAccount(ctx, tx, acc); err != nil {
				return fmt.Errorf("failed to update cash account balance: %w", err)
			}

			// Create cash flow entry
			desc := fmt.Sprintf("Cicilan Pengeluaran %s (%s)", expense.ExpenseNumber, p.PaymentMethod)
			cf := &cfDomain.CashFlow{
				ID:              uuid.New(),
				TransactionDate: p.PaymentDate,
				ReferenceType:   "EXPENSE_PAYMENT",
				ReferenceID:     &p.ID,
				Type:            cfDomain.TypeCredit,
				Amount:          p.Amount,
				PaymentMethod:   p.PaymentMethod,
				Description:     &desc,
				CashierID:       cashierID,
			}
			if err := u.cfRepo.CreateTx(ctx, tx, cf); err != nil {
				return fmt.Errorf("failed to record cash flow: %w", err)
			}
		}

		// Update Status
		newTotalPaid := totalPaidSoFar + totalNewPayment
		newStatus := domain.StatusPartial
		if newTotalPaid >= expense.Amount {
			newStatus = domain.StatusPaid
		}

		if err := u.repo.UpdateExpenseStatusTx(ctx, tx, expense.ID, newStatus); err != nil {
			return fmt.Errorf("failed to update expense status: %w", err)
		}

		return nil
	})
}

func (u *expenseUsecase) FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.Expense, error) {
	return u.repo.FindExpenseByID(ctx, id)
}

func (u *expenseUsecase) FindAllExpenses(ctx context.Context, filter domain.ExpenseFilter) ([]domain.Expense, int64, error) {
	return u.repo.FindAllExpenses(ctx, filter)
}

func (u *expenseUsecase) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		expense, err := u.repo.FindExpenseByIDTx(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("expense not found: %w", err)
		}

		// 1. Restore balances for all payments made
		for _, p := range expense.Payments {
			acc, err := u.cfRepo.FindAccountByNameWithLock(ctx, tx, p.PaymentMethod)
			if err != nil {
				return fmt.Errorf("cash account %s not found: %w", p.PaymentMethod, err)
			}
			acc.Balance += p.Amount
			if err := u.cfRepo.UpdateAccount(ctx, tx, acc); err != nil {
				return fmt.Errorf("failed to restore cash account balance: %w", err)
			}

			// Delete cash flow record
			if err := tx.Where("reference_type = ? AND reference_id = ?", "EXPENSE_PAYMENT", p.ID).Delete(&cfDomain.CashFlow{}).Error; err != nil {
				return fmt.Errorf("failed to delete cash flow: %w", err)
			}
		}

		// 2. Cascade delete will handle Items & Payments deletion since Constraint is ON DELETE CASCADE
		if err := u.repo.DeleteExpenseTx(ctx, tx, expense.ID); err != nil {
			return fmt.Errorf("failed to delete expense: %w", err)
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

func (u *expenseUsecase) GetWidgets(ctx context.Context, filter domain.ExpenseFilter) (*domain.ExpenseWidgetsRes, error) {
	return u.repo.GetWidgets(ctx, filter)
}
