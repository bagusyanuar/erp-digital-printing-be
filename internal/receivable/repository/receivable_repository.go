package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	orderDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/receivable/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type receivableRepository struct {
	db *gorm.DB
}

func NewReceivableRepository(db *gorm.DB) domain.ReceivableRepository {
	return &receivableRepository{db: db}
}

type customerAggRow struct {
	ID                  string  `gorm:"column:id"`
	CustomerName        string  `gorm:"column:customer_name"`
	CustomerPhone       string  `gorm:"column:customer_phone"`
	CustomerType        string  `gorm:"column:customer_type"`
	TotalInvoicesCount  int     `gorm:"column:total_invoices_count"`
	UnpaidInvoicesCount int     `gorm:"column:unpaid_invoices_count"`
	TotalAmount         float64 `gorm:"column:total_amount"`
	TotalPaid           float64 `gorm:"column:total_paid"`
	TotalRemaining      float64 `gorm:"column:total_remaining"`
}

func (r *receivableRepository) FindCustomerReceivables(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	status string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.CustomerReceivable, int64, error) {
	var dateFilterReseller string
	var dateFilterRetail string
	var dateArgs []any

	if startDate != nil && endDate != nil {
		dateFilterReseller = " AND o.created_at >= ? AND o.created_at <= ?"
		dateFilterRetail = " AND o.created_at >= ? AND o.created_at <= ?"
		dateArgs = append(dateArgs, *startDate, *endDate)
	}

	baseQuery := fmt.Sprintf(`
		SELECT 
			r.id::text AS id,
			r.name AS customer_name,
			COALESCE(NULLIF(r.phone, ''), '-') AS customer_phone,
			'RESELLER' AS customer_type,
			COUNT(o.id) AS total_invoices_count,
			COUNT(o.id) AS unpaid_invoices_count,
			COALESCE(SUM(o.grand_total), 0) AS total_amount,
			COALESCE(SUM(o.amount_paid), 0) AS total_paid,
			COALESCE(SUM(o.grand_total - o.amount_paid), 0) AS total_remaining
		FROM orders o
		JOIN resellers r ON o.reseller_id = r.id
		WHERE o.payment_status IN ('UNPAID', 'PARTIAL_PAID')
		  AND o.status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')
		  AND o.invoice_number IS NOT NULL
		  AND o.deleted_at IS NULL
		  %s
		GROUP BY r.id, r.name, r.phone

		UNION ALL

		SELECT 
			CASE 
				WHEN NULLIF(TRIM(o.customer_phone), '') IS NOT NULL 
					THEN 'retail:phone:' || TRIM(o.customer_phone)
				ELSE 'retail:order:' || o.id::text 
			END AS id,
			COALESCE(NULLIF(MAX(o.customer_name), ''), 'Pelanggan Umum') AS customer_name,
			COALESCE(NULLIF(MAX(o.customer_phone), ''), '-') AS customer_phone,
			'RETAIL' AS customer_type,
			COUNT(o.id) AS total_invoices_count,
			COUNT(o.id) AS unpaid_invoices_count,
			COALESCE(SUM(o.grand_total), 0) AS total_amount,
			COALESCE(SUM(o.amount_paid), 0) AS total_paid,
			COALESCE(SUM(o.grand_total - o.amount_paid), 0) AS total_remaining
		FROM orders o
		WHERE o.reseller_id IS NULL
		  AND o.payment_status IN ('UNPAID', 'PARTIAL_PAID')
		  AND o.status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')
		  AND o.invoice_number IS NOT NULL
		  AND o.deleted_at IS NULL
		  %s
		GROUP BY 
			CASE 
				WHEN NULLIF(TRIM(o.customer_phone), '') IS NOT NULL 
					THEN 'retail:phone:' || TRIM(o.customer_phone)
				ELSE 'retail:order:' || o.id::text 
			END
	`, dateFilterReseller, dateFilterRetail)

	allArgs := append([]any{}, dateArgs...)
	allArgs = append(allArgs, dateArgs...)

	tx := r.db.WithContext(ctx).Table("( " + baseQuery + " ) AS c", allArgs...)

	// Filter Search
	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		tx = tx.Where("c.customer_name ILIKE ? OR c.customer_phone ILIKE ?", s, s)
	}

	// Filter Status
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PARTIAL":
		tx = tx.Where("c.total_paid > 0")
	case "UNPAID":
		tx = tx.Where("c.total_paid = 0")
	}

	// Count Total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch Customer Page
	var rows []customerAggRow
	if err := tx.Order("c.total_remaining DESC").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	if len(rows) == 0 {
		return []domain.CustomerReceivable{}, total, nil
	}

	// Step 2: Batch Load Invoices for this page of customers
	var resellerIDs []uuid.UUID
	var retailKeys []string

	for _, row := range rows {
		if row.CustomerType == "RESELLER" {
			if id, err := uuid.Parse(row.ID); err == nil {
				resellerIDs = append(resellerIDs, id)
			}
		} else {
			retailKeys = append(retailKeys, row.ID)
		}
	}

	orderQuery := r.db.WithContext(ctx).
		Preload("OrderPayments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("OrderPayments.Cashier").
		Where("payment_status IN ('UNPAID', 'PARTIAL_PAID')").
		Where("status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')").
		Where("invoice_number IS NOT NULL").
		Where("deleted_at IS NULL")

	if startDate != nil && endDate != nil {
		orderQuery = orderQuery.Where("created_at >= ? AND created_at <= ?", *startDate, *endDate)
	}

	var conditions []string
	var condArgs []any

	if len(resellerIDs) > 0 {
		conditions = append(conditions, "reseller_id IN (?)")
		condArgs = append(condArgs, resellerIDs)
	}
	if len(retailKeys) > 0 {
		conditions = append(conditions, "(CASE WHEN NULLIF(TRIM(customer_phone), '') IS NOT NULL THEN 'retail:phone:' || TRIM(customer_phone) ELSE 'retail:order:' || id::text END) IN (?)")
		condArgs = append(condArgs, retailKeys)
	}

	var orders []orderDomain.Order
	if len(conditions) > 0 {
		orderQuery = orderQuery.Where("("+strings.Join(conditions, " OR ")+")", condArgs...)
		if err := orderQuery.Order("created_at DESC").Find(&orders).Error; err != nil {
			return nil, 0, err
		}
	}

	// Map Invoices to Customer ID
	invoicesByCustID := make(map[string][]domain.ReceivableInvoiceItem)
	for _, o := range orders {
		var custKey string
		if o.ResellerID != nil {
			custKey = o.ResellerID.String()
		} else {
			phone := ""
			if o.CustomerPhone != nil {
				phone = strings.TrimSpace(*o.CustomerPhone)
			}
			if phone != "" {
				custKey = "retail:phone:" + phone
			} else {
				custKey = "retail:order:" + o.ID.String()
			}
		}

		inv := mapOrderToReceivableInvoice(o)
		invoicesByCustID[custKey] = append(invoicesByCustID[custKey], inv)
	}

	// Assemble final result
	customers := make([]domain.CustomerReceivable, len(rows))
	for i, row := range rows {
		invs := invoicesByCustID[row.ID]
		if invs == nil {
			invs = []domain.ReceivableInvoiceItem{}
		}

		customers[i] = domain.CustomerReceivable{
			ID:                  row.ID,
			CustomerName:        row.CustomerName,
			CustomerPhone:       row.CustomerPhone,
			CustomerType:        row.CustomerType,
			TotalInvoicesCount:  row.TotalInvoicesCount,
			UnpaidInvoicesCount: row.UnpaidInvoicesCount,
			TotalAmount:         row.TotalAmount,
			TotalPaid:           row.TotalPaid,
			TotalRemaining:      row.TotalRemaining,
			Invoices:            invs,
		}
	}

	return customers, total, nil
}

func (r *receivableRepository) FindSettledInvoices(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.SettledInvoiceReceivable, int64, error) {
	db := r.db.WithContext(ctx).Model(&orderDomain.Order{}).
		Joins("LEFT JOIN resellers r ON orders.reseller_id = r.id").
		Where("orders.payment_status = ?", orderDomain.PaymentStatusPaid).
		Where("orders.status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')").
		Where("orders.invoice_number IS NOT NULL").
		Where("orders.deleted_at IS NULL")

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.TrimSpace(search) + "%"
		db = db.Where(`(
			orders.invoice_number ILIKE ? OR 
			orders.job_number ILIKE ? OR 
			r.name ILIKE ? OR 
			r.phone ILIKE ? OR 
			orders.customer_name ILIKE ? OR 
			orders.customer_phone ILIKE ?
		)`, s, s, s, s, s, s)
	}

	if startDate != nil && endDate != nil {
		db = db.Where("orders.created_at >= ? AND orders.created_at <= ?", *startDate, *endDate)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []orderDomain.Order
	if err := db.
		Preload("Reseller").
		Preload("OrderPayments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at ASC")
		}).
		Preload("OrderPayments.Cashier").
		Order("orders.created_at DESC").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	settledList := make([]domain.SettledInvoiceReceivable, len(orders))
	for i, o := range orders {
		customerType := "RETAIL"
		custID := "retail:order:" + o.ID.String()
		custName := "Pelanggan Umum"
		custPhone := "-"

		if o.Reseller != nil {
			customerType = "RESELLER"
			custID = o.Reseller.ID.String()
			custName = o.Reseller.Name
			custPhone = o.Reseller.Phone
		} else {
			phone := ""
			if o.CustomerPhone != nil && *o.CustomerPhone != "" {
				phone = strings.TrimSpace(*o.CustomerPhone)
				custPhone = phone
			}
			if o.CustomerName != nil && *o.CustomerName != "" {
				custName = *o.CustomerName
			}
			if phone != "" {
				custID = "retail:phone:" + phone
			} else {
				custID = "retail:order:" + o.ID.String()
			}
		}

		invNum := "-"
		if o.InvoiceNumber != nil && *o.InvoiceNumber != "" {
			invNum = *o.InvoiceNumber
		}

		// Calculate payment method label & settled date
		paymentMethod := "cash"
		uniqueMethods := make(map[string]bool)
		settledDate := o.CreatedAt.Format("2006-01-02 15:04:05")

		payments := make([]domain.ReceivablePaymentItem, len(o.OrderPayments))
		for j, p := range o.OrderPayments {
			uniqueMethods[p.PaymentMethod] = true
			paymentDate := p.CreatedAt.Format("2006-01-02 15:04:05")
			settledDate = paymentDate

			cashierName := "Kasir"
			if p.Cashier != nil {
				cashierName = p.Cashier.Username
			}

			payments[j] = domain.ReceivablePaymentItem{
				ID:                  p.ID,
				ReceivableInvoiceID: o.ID,
				Amount:              p.Amount,
				PaymentMethod:       p.PaymentMethod,
				CashierName:         cashierName,
				PaymentDate:         paymentDate,
			}
		}

		if len(uniqueMethods) > 1 {
			paymentMethod = "split"
		} else if len(uniqueMethods) == 1 {
			for m := range uniqueMethods {
				paymentMethod = m
			}
		}

		settledList[i] = domain.SettledInvoiceReceivable{
			ID:            o.ID,
			CustomerID:    custID,
			CustomerName:  custName,
			CustomerPhone: custPhone,
			CustomerType:  customerType,
			InvoiceNumber: invNum,
			JobNumber:     o.JobNumber,
			OrderDate:     o.CreatedAt.Format("2006-01-02 15:04:05"),
			SettledDate:   settledDate,
			TotalAmount:   o.GrandTotal,
			AmountPaid:    o.AmountPaid,
			PaymentMethod: paymentMethod,
			Payments:      payments,
			Notes:         o.Notes,
		}
	}

	return settledList, total, nil
}

func (r *receivableRepository) GetSummary(ctx context.Context) (*domain.CustomerReceivableSummary, error) {
	res := &domain.CustomerReceivableSummary{}

	// 1. Total Active Receivable
	err := r.db.WithContext(ctx).
		Model(&orderDomain.Order{}).
		Select("COALESCE(SUM(grand_total - amount_paid), 0)").
		Where("payment_status IN ('UNPAID', 'PARTIAL_PAID')").
		Where("status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')").
		Where("invoice_number IS NOT NULL").
		Where("deleted_at IS NULL").
		Scan(&res.TotalActiveReceivable).Error
	if err != nil {
		return nil, err
	}

	// 2. Total Settled Receivable (from REPAYMENT order_payments)
	err = r.db.WithContext(ctx).
		Model(&orderDomain.OrderPayment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("payment_type = 'REPAYMENT'").
		Where("deleted_at IS NULL").
		Scan(&res.TotalSettledReceivable).Error
	if err != nil {
		return nil, err
	}

	// 3. Active Customer Count
	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT r.id::text
			FROM orders o
			JOIN resellers r ON o.reseller_id = r.id
			WHERE o.payment_status IN ('UNPAID', 'PARTIAL_PAID')
			  AND o.status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')
			  AND o.invoice_number IS NOT NULL
			  AND o.deleted_at IS NULL
			GROUP BY r.id

			UNION

			SELECT 
				CASE 
					WHEN NULLIF(TRIM(o.customer_phone), '') IS NOT NULL 
						THEN 'retail:phone:' || TRIM(o.customer_phone)
					ELSE 'retail:order:' || o.id::text 
				END
			FROM orders o
			WHERE o.reseller_id IS NULL
			  AND o.payment_status IN ('UNPAID', 'PARTIAL_PAID')
			  AND o.status NOT IN ('DRAFT', 'PENDING_PAYMENT', 'CANCELLED', 'REFUND')
			  AND o.invoice_number IS NOT NULL
			  AND o.deleted_at IS NULL
			GROUP BY 
				CASE 
					WHEN NULLIF(TRIM(o.customer_phone), '') IS NOT NULL 
						THEN 'retail:phone:' || TRIM(o.customer_phone)
					ELSE 'retail:order:' || o.id::text 
				END
		) AS active_cust
	`
	err = r.db.WithContext(ctx).Raw(countQuery).Scan(&res.ActiveCustomerCount).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func mapOrderToReceivableInvoice(o orderDomain.Order) domain.ReceivableInvoiceItem {
	invNum := "-"
	if o.InvoiceNumber != nil && *o.InvoiceNumber != "" {
		invNum = *o.InvoiceNumber
	}

	remaining := o.GrandTotal - o.AmountPaid
	if remaining < 0 {
		remaining = 0
	}

	payments := make([]domain.ReceivablePaymentItem, len(o.OrderPayments))
	for i, p := range o.OrderPayments {
		cashierName := "Kasir"
		if p.Cashier != nil {
			cashierName = p.Cashier.Username
		}
		payments[i] = domain.ReceivablePaymentItem{
			ID:                  p.ID,
			ReceivableInvoiceID: o.ID,
			Amount:              p.Amount,
			PaymentMethod:       p.PaymentMethod,
			CashierName:         cashierName,
			PaymentDate:         p.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return domain.ReceivableInvoiceItem{
		ID:              o.ID,
		InvoiceNumber:   invNum,
		JobNumber:       o.JobNumber,
		OrderDate:       o.CreatedAt.Format("2006-01-02 15:04:05"),
		TotalAmount:     o.GrandTotal,
		AmountPaid:      o.AmountPaid,
		RemainingAmount: remaining,
		PaymentStatus:   o.PaymentStatus,
		Notes:           o.Notes,
		Payments:        payments,
	}
}
