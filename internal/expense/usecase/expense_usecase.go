package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cfDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
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

func (u *expenseUsecase) ReorderCategories(ctx context.Context, items []request.ReorderItem) error {
	return u.repo.ReorderCategories(ctx, items)
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

func (u *expenseUsecase) GetCategoryExpenses(ctx context.Context, params request.PaginationParam, search string, startDate *time.Time, endDate *time.Time, productCategoryID *uuid.UUID, group string, status string) ([]domain.CategoryExpenseItem, int64, error) {
	return u.repo.GetCategoryExpenses(ctx, params, search, startDate, endDate, productCategoryID, group, status)
}

func (u *expenseUsecase) ExportExcel(ctx context.Context, filter domain.ExpenseFilter) (*excelize.File, error) {
	// 1. Fetch expenses without pagination limit
	filter.Page = 1
	filter.Limit = 100000
	expenses, _, err := u.repo.FindAllExpenses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}

	f := excelize.NewFile()

	sheet1 := "Pengeluaran"
	sheet2 := "Detail Item Pengeluaran"

	_ = f.SetSheetName("Sheet1", sheet1)
	_, err = f.NewSheet(sheet2)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(0)

	// Period text
	var periodeText string
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.Format("2006-01-02") == filter.EndDate.Format("2006-01-02") {
			periodeText = fmt.Sprintf("Periode: %s", filter.StartDate.Format("02/01/2006"))
		} else {
			periodeText = fmt.Sprintf("Periode: %s s/d %s", filter.StartDate.Format("02/01/2006"), filter.EndDate.Format("02/01/2006"))
		}
	} else if filter.StartDate != nil {
		periodeText = fmt.Sprintf("Periode: Mulai %s", filter.StartDate.Format("02/01/2006"))
	} else if filter.EndDate != nil {
		periodeText = fmt.Sprintf("Periode: Sampai %s", filter.EndDate.Format("02/01/2006"))
	} else {
		periodeText = "Periode: Semua Periode"
	}

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1E3A8A", Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	periodStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: false, Color: "4B5563", Size: 10, Italic: true},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E3A8A"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D1D5DB", Style: 1},
			{Type: "bottom", Color: "D1D5DB", Style: 1},
			{Type: "left", Color: "D1D5DB", Style: 1},
			{Type: "right", Color: "D1D5DB", Style: 1},
		},
	})

	textLeftStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	textCenterStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	numFmt := "#,##0"
	numRightStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	totalLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1F2937"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F3F4F6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "1F2937", Style: 1},
			{Type: "bottom", Color: "1F2937", Style: 6},
			{Type: "left", Color: "1F2937", Style: 1},
			{Type: "right", Color: "1F2937", Style: 1},
		},
	})

	totalNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true, Color: "1F2937"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"F3F4F6"}, Pattern: 1},
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "1F2937", Style: 1},
			{Type: "bottom", Color: "1F2937", Style: 6},
			{Type: "left", Color: "1F2937", Style: 1},
			{Type: "right", Color: "1F2937", Style: 1},
		},
	})

	paidCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "065F46"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D1FAE5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "A7F3D0", Style: 1},
			{Type: "bottom", Color: "A7F3D0", Style: 1},
			{Type: "left", Color: "A7F3D0", Style: 1},
			{Type: "right", Color: "A7F3D0", Style: 1},
		},
	})

	partialCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "92400E"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF3C7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FDE68A", Style: 1},
			{Type: "bottom", Color: "FDE68A", Style: 1},
			{Type: "left", Color: "FDE68A", Style: 1},
			{Type: "right", Color: "FDE68A", Style: 1},
		},
	})

	unpaidCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "991B1B"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEE2E2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FCA5A5", Style: 1},
			{Type: "bottom", Color: "FCA5A5", Style: 1},
			{Type: "left", Color: "FCA5A5", Style: 1},
			{Type: "right", Color: "FCA5A5", Style: 1},
		},
	})

	// Sheet 1: Setup Header
	_ = f.SetRowHeight(sheet1, 1, 24)
	_ = f.SetRowHeight(sheet1, 2, 18)
	_ = f.SetRowHeight(sheet1, 4, 26)

	_ = f.SetCellValue(sheet1, "A1", "LAPORAN PENGELUARAN")
	_ = f.SetCellStyle(sheet1, "A1", "A1", titleStyle)
	_ = f.SetCellValue(sheet1, "A2", periodeText)
	_ = f.SetCellStyle(sheet1, "A2", "A2", periodStyle)

	sheet1Headers := []string{
		"No.", "No. Pengeluaran", "No. Invoice", "Tanggal", "Vendor / Supplier",
		"Status Bayar", "Metode Bayar", "Keterangan", "Total Nominal", "Terbayar", "Sisa Hutang",
	}

	for i, h := range sheet1Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet1, cell, h)
		_ = f.SetCellStyle(sheet1, cell, cell, headerStyle)
	}

	_ = f.SetColWidth(sheet1, "A", "A", 6)
	_ = f.SetColWidth(sheet1, "B", "C", 22)
	_ = f.SetColWidth(sheet1, "D", "D", 16)
	_ = f.SetColWidth(sheet1, "E", "E", 26)
	_ = f.SetColWidth(sheet1, "F", "F", 16)
	_ = f.SetColWidth(sheet1, "G", "G", 18)
	_ = f.SetColWidth(sheet1, "H", "H", 30)
	_ = f.SetColWidth(sheet1, "I", "K", 18)

	// Sheet 2: Setup Header
	_ = f.SetRowHeight(sheet2, 1, 24)
	_ = f.SetRowHeight(sheet2, 2, 18)
	_ = f.SetRowHeight(sheet2, 4, 26)

	_ = f.SetCellValue(sheet2, "A1", "DETAIL ITEM PENGELUARAN")
	_ = f.SetCellStyle(sheet2, "A1", "A1", titleStyle)
	_ = f.SetCellValue(sheet2, "A2", periodeText)
	_ = f.SetCellStyle(sheet2, "A2", "A2", periodStyle)

	sheet2Headers := []string{
		"No.", "No. Pengeluaran", "Tanggal", "Vendor / Supplier",
		"Grup Biaya", "Kategori Biaya", "Kategori Produk", "Deskripsi Item", "Nominal",
	}

	for i, h := range sheet2Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet2, cell, h)
		_ = f.SetCellStyle(sheet2, cell, cell, headerStyle)
	}

	_ = f.SetColWidth(sheet2, "A", "A", 6)
	_ = f.SetColWidth(sheet2, "B", "B", 22)
	_ = f.SetColWidth(sheet2, "C", "C", 16)
	_ = f.SetColWidth(sheet2, "D", "D", 26)
	_ = f.SetColWidth(sheet2, "E", "E", 18)
	_ = f.SetColWidth(sheet2, "F", "F", 24)
	_ = f.SetColWidth(sheet2, "G", "G", 20)
	_ = f.SetColWidth(sheet2, "H", "H", 32)
	_ = f.SetColWidth(sheet2, "I", "I", 18)

	var grandTotalExpense, grandTotalPaid, grandTotalRemaining float64
	var grandTotalItemAmount float64

	itemRowIdx := 5
	itemSeq := 1

	for i, exp := range expenses {
		expRowIdx := i + 5

		inv := "-"
		if exp.InvoiceNumber != nil && *exp.InvoiceNumber != "" {
			inv = *exp.InvoiceNumber
		}

		vendor := exp.VendorName
		if vendor == "" && exp.Supplier != nil && exp.Supplier.Name != "" {
			vendor = exp.Supplier.Name
		} else if exp.Supplier != nil && exp.Supplier.Name != "" && exp.Supplier.Name != vendor {
			vendor = fmt.Sprintf("%s (%s)", vendor, exp.Supplier.Name)
		}
		if vendor == "" {
			vendor = "-"
		}

		desc := "-"
		if exp.Description != nil && *exp.Description != "" {
			desc = *exp.Description
		}

		var totalPaid float64
		var methods []string
		for _, p := range exp.Payments {
			totalPaid += p.Amount
			if p.PaymentMethod != "" {
				found := false
				for _, m := range methods {
					if m == p.PaymentMethod {
						found = true
						break
					}
				}
				if !found {
					methods = append(methods, p.PaymentMethod)
				}
			}
		}

		if exp.Status == domain.StatusPaid && totalPaid == 0 {
			totalPaid = exp.Amount
		}

		remaining := exp.Amount - totalPaid
		if remaining < 0 {
			remaining = 0
		}

		paymentMethodStr := "-"
		if len(methods) > 0 {
			paymentMethodStr = strings.Join(methods, ", ")
		}

		statusText := "Belum Lunas"
		if exp.Status == domain.StatusPaid {
			statusText = "Lunas"
		}

		grandTotalExpense += exp.Amount
		grandTotalPaid += totalPaid
		grandTotalRemaining += remaining

		// Sheet 1 row
		_ = f.SetRowHeight(sheet1, expRowIdx, 20)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", expRowIdx), i+1)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("B%d", expRowIdx), exp.ExpenseNumber)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("C%d", expRowIdx), inv)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("D%d", expRowIdx), exp.ExpenseDate.Format("02/01/2006 15:04"))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("E%d", expRowIdx), vendor)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("F%d", expRowIdx), statusText)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("G%d", expRowIdx), paymentMethodStr)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("H%d", expRowIdx), desc)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("I%d", expRowIdx), exp.Amount)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("J%d", expRowIdx), totalPaid)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("K%d", expRowIdx), remaining)

		_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", expRowIdx), fmt.Sprintf("A%d", expRowIdx), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("B%d", expRowIdx), fmt.Sprintf("C%d", expRowIdx), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("D%d", expRowIdx), fmt.Sprintf("D%d", expRowIdx), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("E%d", expRowIdx), fmt.Sprintf("E%d", expRowIdx), textLeftStyle)

		statusCell := fmt.Sprintf("F%d", expRowIdx)
		switch exp.Status {
		case domain.StatusPaid:
			_ = f.SetCellStyle(sheet1, statusCell, statusCell, paidCenterStyle)
		case domain.StatusPartial:
			_ = f.SetCellStyle(sheet1, statusCell, statusCell, partialCenterStyle)
		case domain.StatusUnpaid:
			_ = f.SetCellStyle(sheet1, statusCell, statusCell, unpaidCenterStyle)
		default:
			_ = f.SetCellStyle(sheet1, statusCell, statusCell, textCenterStyle)
		}

		_ = f.SetCellStyle(sheet1, fmt.Sprintf("G%d", expRowIdx), fmt.Sprintf("G%d", expRowIdx), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("H%d", expRowIdx), fmt.Sprintf("H%d", expRowIdx), textLeftStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("I%d", expRowIdx), fmt.Sprintf("K%d", expRowIdx), numRightStyle)

		// Sheet 2 rows (Items)
		for _, item := range exp.Items {
			_ = f.SetRowHeight(sheet2, itemRowIdx, 20)

			groupName := item.ExpenseCategory.Group
			if groupName == domain.GroupProduction {
				groupName = "PRODUKSI"
			} else if groupName == domain.GroupOperational {
				groupName = "OPERASIONAL"
			}

			catName := item.ExpenseCategory.Name
			if catName == "" {
				catName = "-"
			}

			prodCatName := "-"
			if item.ExpenseCategory.ProductCategory != nil && item.ExpenseCategory.ProductCategory.Name != "" {
				prodCatName = item.ExpenseCategory.ProductCategory.Name
			}

			itemDesc := "-"
			if item.Description != nil && *item.Description != "" {
				itemDesc = *item.Description
			}

			grandTotalItemAmount += item.Amount

			_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", itemRowIdx), itemSeq)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("B%d", itemRowIdx), exp.ExpenseNumber)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("C%d", itemRowIdx), exp.ExpenseDate.Format("02/01/2006 15:04"))
			_ = f.SetCellValue(sheet2, fmt.Sprintf("D%d", itemRowIdx), vendor)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("E%d", itemRowIdx), groupName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("F%d", itemRowIdx), catName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("G%d", itemRowIdx), prodCatName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("H%d", itemRowIdx), itemDesc)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("I%d", itemRowIdx), item.Amount)

			_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", itemRowIdx), fmt.Sprintf("A%d", itemRowIdx), textCenterStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("B%d", itemRowIdx), fmt.Sprintf("C%d", itemRowIdx), textCenterStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("D%d", itemRowIdx), fmt.Sprintf("D%d", itemRowIdx), textLeftStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("E%d", itemRowIdx), fmt.Sprintf("G%d", itemRowIdx), textCenterStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("H%d", itemRowIdx), fmt.Sprintf("H%d", itemRowIdx), textLeftStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("I%d", itemRowIdx), fmt.Sprintf("I%d", itemRowIdx), numRightStyle)

			itemRowIdx++
			itemSeq++
		}
	}

	// Total Row Sheet 1
	totRow1 := len(expenses) + 5
	_ = f.SetRowHeight(sheet1, totRow1, 24)
	_ = f.MergeCell(sheet1, fmt.Sprintf("A%d", totRow1), fmt.Sprintf("H%d", totRow1))
	_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", totRow1), "TOTAL")
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", totRow1), fmt.Sprintf("H%d", totRow1), totalLabelStyle)

	_ = f.SetCellValue(sheet1, fmt.Sprintf("I%d", totRow1), grandTotalExpense)
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("I%d", totRow1), fmt.Sprintf("I%d", totRow1), totalNumStyle)

	_ = f.SetCellValue(sheet1, fmt.Sprintf("J%d", totRow1), grandTotalPaid)
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("J%d", totRow1), fmt.Sprintf("J%d", totRow1), totalNumStyle)

	_ = f.SetCellValue(sheet1, fmt.Sprintf("K%d", totRow1), grandTotalRemaining)
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("K%d", totRow1), fmt.Sprintf("K%d", totRow1), totalNumStyle)

	// Total Row Sheet 2
	totRow2 := itemRowIdx
	_ = f.SetRowHeight(sheet2, totRow2, 24)
	_ = f.MergeCell(sheet2, fmt.Sprintf("A%d", totRow2), fmt.Sprintf("H%d", totRow2))
	_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", totRow2), "TOTAL")
	_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", totRow2), fmt.Sprintf("H%d", totRow2), totalLabelStyle)

	_ = f.SetCellValue(sheet2, fmt.Sprintf("I%d", totRow2), grandTotalItemAmount)
	_ = f.SetCellStyle(sheet2, fmt.Sprintf("I%d", totRow2), fmt.Sprintf("I%d", totRow2), totalNumStyle)

	return f, nil
}
