package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	expenseDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/payable/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type payableRepository struct {
	db *gorm.DB
}

func NewPayableRepository(db *gorm.DB) domain.PayableRepository {
	return &payableRepository{db: db}
}

type vendorAggRow struct {
	ID                  string  `gorm:"column:id"`
	VendorName          string  `gorm:"column:vendor_name"`
	VendorPhone         string  `gorm:"column:vendor_phone"`
	VendorCategory      string  `gorm:"column:vendor_category"`
	TotalExpensesCount  int     `gorm:"column:total_expenses_count"`
	UnpaidExpensesCount int     `gorm:"column:unpaid_expenses_count"`
	TotalAmount         float64 `gorm:"column:total_amount"`
	TotalPaid           float64 `gorm:"column:total_paid"`
	TotalRemaining      float64 `gorm:"column:total_remaining"`
}

func (r *payableRepository) FindVendorPayables(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	status string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.VendorPayable, int64, error) {
	var dateFilterSupplier string
	var dateFilterGeneral string
	var dateArgs []any

	if startDate != nil && endDate != nil {
		dateFilterSupplier = " AND e.expense_date >= ? AND e.expense_date <= ?"
		dateFilterGeneral = " AND e.expense_date >= ? AND e.expense_date <= ?"
		dateArgs = append(dateArgs, *startDate, *endDate)
	}

	baseQuery := fmt.Sprintf(`
		SELECT 
			s.id::text AS id,
			s.name AS vendor_name,
			COALESCE(NULLIF(s.phone, ''), '-') AS vendor_phone,
			'SUPPLIER' AS vendor_category,
			COUNT(e.id) AS total_expenses_count,
			COUNT(e.id) AS unpaid_expenses_count,
			COALESCE(SUM(e.amount), 0) AS total_amount,
			COALESCE(SUM(COALESCE(ep_sub.total_paid, 0)), 0) AS total_paid,
			COALESCE(SUM(e.amount - COALESCE(ep_sub.total_paid, 0)), 0) AS total_remaining
		FROM expenses e
		JOIN suppliers s ON e.supplier_id = s.id
		LEFT JOIN (
			SELECT expense_id, SUM(amount) AS total_paid
			FROM expense_payments
			WHERE deleted_at IS NULL
			GROUP BY expense_id
		) ep_sub ON ep_sub.expense_id = e.id
		WHERE e.status IN ('UNPAID', 'PARTIAL')
		  AND e.deleted_at IS NULL
		  %s
		GROUP BY s.id, s.name, s.phone

		UNION ALL

		SELECT 
			'general:' || LOWER(TRIM(e.vendor_name)) AS id,
			e.vendor_name AS vendor_name,
			'-' AS vendor_phone,
			'GENERAL' AS vendor_category,
			COUNT(e.id) AS total_expenses_count,
			COUNT(e.id) AS unpaid_expenses_count,
			COALESCE(SUM(e.amount), 0) AS total_amount,
			COALESCE(SUM(COALESCE(ep_sub.total_paid, 0)), 0) AS total_paid,
			COALESCE(SUM(e.amount - COALESCE(ep_sub.total_paid, 0)), 0) AS total_remaining
		FROM expenses e
		LEFT JOIN (
			SELECT expense_id, SUM(amount) AS total_paid
			FROM expense_payments
			WHERE deleted_at IS NULL
			GROUP BY expense_id
		) ep_sub ON ep_sub.expense_id = e.id
		WHERE e.supplier_id IS NULL
		  AND e.status IN ('UNPAID', 'PARTIAL')
		  AND e.deleted_at IS NULL
		  %s
		GROUP BY LOWER(TRIM(e.vendor_name)), e.vendor_name
	`, dateFilterSupplier, dateFilterGeneral)

	allArgs := append([]any{}, dateArgs...)
	allArgs = append(allArgs, dateArgs...)

	tx := r.db.WithContext(ctx).Table("( " + baseQuery + " ) AS v", allArgs...)

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		tx = tx.Where("v.vendor_name ILIKE ? OR v.vendor_phone ILIKE ?", s, s)
	}

	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PARTIAL":
		tx = tx.Where("v.total_paid > 0")
	case "UNPAID":
		tx = tx.Where("v.total_paid = 0")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []vendorAggRow
	if err := tx.Order("v.total_remaining DESC").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	if len(rows) == 0 {
		return []domain.VendorPayable{}, total, nil
	}

	var supplierIDs []uuid.UUID
	var generalVendorKeys []string

	for _, row := range rows {
		if strings.HasPrefix(row.ID, "general:") {
			name := strings.TrimPrefix(row.ID, "general:")
			generalVendorKeys = append(generalVendorKeys, name)
		} else {
			if id, err := uuid.Parse(row.ID); err == nil {
				supplierIDs = append(supplierIDs, id)
			}
		}
	}

	expenseQuery := r.db.WithContext(ctx).
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date ASC")
		}).
		Preload("Payments.Cashier").
		Preload("Items.ExpenseCategory").
		Where("status IN ('UNPAID', 'PARTIAL')").
		Where("deleted_at IS NULL")

	if startDate != nil && endDate != nil {
		expenseQuery = expenseQuery.Where("expense_date >= ? AND expense_date <= ?", *startDate, *endDate)
	}

	var conditions []string
	var condArgs []any

	if len(supplierIDs) > 0 {
		conditions = append(conditions, "supplier_id IN (?)")
		condArgs = append(condArgs, supplierIDs)
	}
	if len(generalVendorKeys) > 0 {
		conditions = append(conditions, "(supplier_id IS NULL AND LOWER(TRIM(vendor_name)) IN (?))")
		condArgs = append(condArgs, generalVendorKeys)
	}

	var expenses []expenseDomain.Expense
	if len(conditions) > 0 {
		expenseQuery = expenseQuery.Where("("+strings.Join(conditions, " OR ")+")", condArgs...)
		if err := expenseQuery.Order("expense_date DESC").Find(&expenses).Error; err != nil {
			return nil, 0, err
		}
	}

	expensesByVendorID := make(map[string][]domain.PayableExpenseItem)
	for _, e := range expenses {
		var vendorKey string
		if e.SupplierID != nil {
			vendorKey = e.SupplierID.String()
		} else {
			vendorKey = "general:" + strings.ToLower(strings.TrimSpace(e.VendorName))
		}

		item := mapExpenseToPayableItem(e)
		expensesByVendorID[vendorKey] = append(expensesByVendorID[vendorKey], item)
	}

	vendors := make([]domain.VendorPayable, len(rows))
	for i, row := range rows {
		exps := expensesByVendorID[row.ID]
		if exps == nil {
			exps = []domain.PayableExpenseItem{}
		}

		vendors[i] = domain.VendorPayable{
			ID:                  row.ID,
			VendorName:          row.VendorName,
			VendorPhone:         row.VendorPhone,
			VendorCategory:      row.VendorCategory,
			TotalExpensesCount:  row.TotalExpensesCount,
			UnpaidExpensesCount: row.UnpaidExpensesCount,
			TotalAmount:         row.TotalAmount,
			TotalPaid:           row.TotalPaid,
			TotalRemaining:      row.TotalRemaining,
			Expenses:            exps,
		}
	}

	return vendors, total, nil
}

func (r *payableRepository) FindSettledPayables(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.SettledExpensePayable, int64, error) {
	db := r.db.WithContext(ctx).Model(&expenseDomain.Expense{}).
		Joins("LEFT JOIN suppliers s ON expenses.supplier_id = s.id").
		Where("expenses.status = ?", expenseDomain.StatusPaid).
		Where("expenses.deleted_at IS NULL")

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		db = db.Where(`(
			expenses.expense_number ILIKE ? OR 
			expenses.invoice_number ILIKE ? OR 
			expenses.vendor_name ILIKE ? OR 
			s.name ILIKE ? OR 
			s.phone ILIKE ?
		)`, s, s, s, s, s)
	}

	if startDate != nil && endDate != nil {
		db = db.Where("expenses.expense_date >= ? AND expenses.expense_date <= ?", *startDate, *endDate)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var expenses []expenseDomain.Expense
	if err := db.
		Preload("Supplier").
		Preload("Payments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("payment_date ASC")
		}).
		Preload("Payments.Cashier").
		Preload("Items.ExpenseCategory").
		Order("expenses.expense_date DESC").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Find(&expenses).Error; err != nil {
		return nil, 0, err
	}

	settledList := make([]domain.SettledExpensePayable, len(expenses))
	for i, e := range expenses {
		vendorID := "general:" + strings.ToLower(strings.TrimSpace(e.VendorName))
		vendorName := e.VendorName
		vendorPhone := "-"
		vendorCategory := "GENERAL"

		if e.Supplier != nil {
			vendorID = e.Supplier.ID.String()
			vendorName = e.Supplier.Name
			if e.Supplier.Phone != "" {
				vendorPhone = e.Supplier.Phone
			}
			vendorCategory = "SUPPLIER"
		}

		paymentMethod := "cash"
		uniqueMethods := make(map[string]bool)
		settledDate := e.ExpenseDate.Format("2006-01-02 15:04:05")
		var totalAmountPaid float64

		payments := make([]domain.PayablePaymentItem, len(e.Payments))
		for j, p := range e.Payments {
			uniqueMethods[strings.ToLower(p.PaymentMethod)] = true
			pDate := p.PaymentDate.Format("2006-01-02 15:04:05")
			settledDate = pDate
			totalAmountPaid += p.Amount

			recordedBy := "Finance"
			if p.Cashier != nil {
				recordedBy = p.Cashier.Username
			}

			payments[j] = domain.PayablePaymentItem{
				ID:               p.ID,
				PayableExpenseID: p.ExpenseID,
				Amount:           p.Amount,
				PaymentMethod:    p.PaymentMethod,
				RecordedBy:       recordedBy,
				PaymentDate:      pDate,
			}
		}

		if len(uniqueMethods) > 1 {
			paymentMethod = "split"
		} else if len(uniqueMethods) == 1 {
			for m := range uniqueMethods {
				paymentMethod = m
			}
		}

		var catNames []string
		seenCat := make(map[string]bool)
		var itemDescs []string
		for _, it := range e.Items {
			cName := it.ExpenseCategory.Name
			if cName != "" && !seenCat[cName] {
				seenCat[cName] = true
				catNames = append(catNames, cName)
			}
			if it.Description != nil && strings.TrimSpace(*it.Description) != "" {
				itemDescs = append(itemDescs, strings.TrimSpace(*it.Description))
			}
		}
		categoryName := strings.Join(catNames, ", ")
		if categoryName == "" {
			categoryName = "Pengeluaran Operasional"
		}

		settledNotes := e.Description
		if (settledNotes == nil || strings.TrimSpace(*settledNotes) == "") && len(itemDescs) > 0 {
			joined := strings.Join(itemDescs, ", ")
			settledNotes = &joined
		}

		settledList[i] = domain.SettledExpensePayable{
			ID:              e.ID,
			VendorID:        vendorID,
			VendorName:      vendorName,
			VendorPhone:     vendorPhone,
			VendorCategory:  vendorCategory,
			ExpenseNumber:   e.ExpenseNumber,
			ReferenceNumber: e.InvoiceNumber,
			CategoryName:    categoryName,
			ExpenseDate:     e.ExpenseDate.Format("2006-01-02 15:04:05"),
			SettledDate:     settledDate,
			TotalAmount:     e.Amount,
			AmountPaid:      totalAmountPaid,
			PaymentMethod:   paymentMethod,
			Payments:        payments,
			Notes:           settledNotes,
		}
	}

	return settledList, total, nil
}

func (r *payableRepository) GetSummary(ctx context.Context) (*domain.VendorPayableSummary, error) {
	res := &domain.VendorPayableSummary{}

	// 1. Total Active Payable
	activeQuery := `
		SELECT COALESCE(SUM(e.amount - COALESCE(ep_sub.total_paid, 0)), 0)
		FROM expenses e
		LEFT JOIN (
			SELECT expense_id, SUM(amount) AS total_paid
			FROM expense_payments
			WHERE deleted_at IS NULL
			GROUP BY expense_id
		) ep_sub ON ep_sub.expense_id = e.id
		WHERE e.status IN ('UNPAID', 'PARTIAL')
		  AND e.deleted_at IS NULL
	`
	if err := r.db.WithContext(ctx).Raw(activeQuery).Scan(&res.TotalActivePayable).Error; err != nil {
		return nil, err
	}

	// 2. Total Settled Payable
	if err := r.db.WithContext(ctx).
		Model(&expenseDomain.Expense{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status = ?", expenseDomain.StatusPaid).
		Where("deleted_at IS NULL").
		Scan(&res.TotalSettledPayable).Error; err != nil {
		return nil, err
	}

	// 3. Active Vendor Count
	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT s.id::text
			FROM expenses e
			JOIN suppliers s ON e.supplier_id = s.id
			WHERE e.status IN ('UNPAID', 'PARTIAL')
			  AND e.deleted_at IS NULL
			GROUP BY s.id

			UNION

			SELECT 'general:' || LOWER(TRIM(e.vendor_name))
			FROM expenses e
			WHERE e.supplier_id IS NULL
			  AND e.status IN ('UNPAID', 'PARTIAL')
			  AND e.deleted_at IS NULL
			GROUP BY LOWER(TRIM(e.vendor_name))
		) AS active_v
	`
	if err := r.db.WithContext(ctx).Raw(countQuery).Scan(&res.ActiveVendorCount).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (r *payableRepository) FindVendorPayableByExpenseID(ctx context.Context, expenseID uuid.UUID) (*domain.VendorPayable, error) {
	var exp expenseDomain.Expense
	if err := r.db.WithContext(ctx).
		Preload("Supplier").
		First(&exp, "id = ?", expenseID).Error; err != nil {
		return nil, err
	}

	var vendorID string
	var vendorName string
	var vendorPhone string
	var vendorCategory string

	var queryFilter *gorm.DB
	if exp.SupplierID != nil {
		vendorID = exp.SupplierID.String()
		vendorName = exp.VendorName
		vendorPhone = "-"
		if exp.Supplier != nil {
			vendorName = exp.Supplier.Name
			if exp.Supplier.Phone != "" {
				vendorPhone = exp.Supplier.Phone
			}
		}
		vendorCategory = "SUPPLIER"
		queryFilter = r.db.WithContext(ctx).Where("supplier_id = ?", exp.SupplierID)
	} else {
		vendorID = "general:" + strings.ToLower(strings.TrimSpace(exp.VendorName))
		vendorName = exp.VendorName
		vendorPhone = "-"
		vendorCategory = "GENERAL"
		queryFilter = r.db.WithContext(ctx).Where("supplier_id IS NULL AND LOWER(TRIM(vendor_name)) = ?", strings.ToLower(strings.TrimSpace(exp.VendorName)))
	}

	var relatedExpenses []expenseDomain.Expense
	if err := queryFilter.
		Preload("Payments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("payment_date ASC")
		}).
		Preload("Payments.Cashier").
		Preload("Items.ExpenseCategory").
		Where("deleted_at IS NULL").
		Order("expense_date DESC").
		Find(&relatedExpenses).Error; err != nil {
		return nil, err
	}

	var totalAmount float64
	var totalPaid float64
	var totalRemaining float64
	var unpaidCount int
	expenseItems := make([]domain.PayableExpenseItem, len(relatedExpenses))

	for i, e := range relatedExpenses {
		item := mapExpenseToPayableItem(e)
		expenseItems[i] = item

		totalAmount += item.TotalAmount
		totalPaid += item.AmountPaid
		totalRemaining += item.RemainingAmount

		if item.PaymentStatus != expenseDomain.StatusPaid {
			unpaidCount++
		}
	}

	return &domain.VendorPayable{
		ID:                  vendorID,
		VendorName:          vendorName,
		VendorPhone:         vendorPhone,
		VendorCategory:      vendorCategory,
		TotalExpensesCount:  len(relatedExpenses),
		UnpaidExpensesCount: unpaidCount,
		TotalAmount:         totalAmount,
		TotalPaid:           totalPaid,
		TotalRemaining:      totalRemaining,
		Expenses:            expenseItems,
	}, nil
}

func mapExpenseToPayableItem(e expenseDomain.Expense) domain.PayableExpenseItem {
	var amountPaid float64
	payments := make([]domain.PayablePaymentItem, len(e.Payments))
	for i, p := range e.Payments {
		amountPaid += p.Amount
		recordedBy := "Finance"
		if p.Cashier != nil {
			recordedBy = p.Cashier.Username
		}
		payments[i] = domain.PayablePaymentItem{
			ID:               p.ID,
			PayableExpenseID: p.ExpenseID,
			Amount:           p.Amount,
			PaymentMethod:    p.PaymentMethod,
			RecordedBy:       recordedBy,
			PaymentDate:      p.PaymentDate.Format("2006-01-02 15:04:05"),
		}
	}

	remaining := e.Amount - amountPaid
	if remaining < 0 {
		remaining = 0
	}

	var catNames []string
	seenCat := make(map[string]bool)
	var itemDescs []string
	for _, it := range e.Items {
		cName := it.ExpenseCategory.Name
		if cName != "" && !seenCat[cName] {
			seenCat[cName] = true
			catNames = append(catNames, cName)
		}
		if it.Description != nil && strings.TrimSpace(*it.Description) != "" {
			itemDescs = append(itemDescs, strings.TrimSpace(*it.Description))
		}
	}
	categoryName := strings.Join(catNames, ", ")
	if categoryName == "" {
		categoryName = "Pengeluaran Operasional"
	}

	expenseNotes := e.Description
	if (expenseNotes == nil || strings.TrimSpace(*expenseNotes) == "") && len(itemDescs) > 0 {
		joined := strings.Join(itemDescs, ", ")
		expenseNotes = &joined
	}

	return domain.PayableExpenseItem{
		ID:              e.ID,
		ExpenseNumber:   e.ExpenseNumber,
		ReferenceNumber: e.InvoiceNumber,
		CategoryName:    categoryName,
		ExpenseDate:     e.ExpenseDate.Format("2006-01-02 15:04:05"),
		TotalAmount:     e.Amount,
		AmountPaid:      amountPaid,
		RemainingAmount: remaining,
		PaymentStatus:   e.Status,
		Notes:           expenseNotes,
		Payments:        payments,
	}
}

