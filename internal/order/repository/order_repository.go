package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Reseller").
		Preload("Designer").
		Preload("Cashier").
		Preload("OrderItems").
		Preload("OrderItems.ProductVariant").
		Preload("OrderItems.ProductVariant.Product").
		Preload("OrderItems.Finishings").
		Preload("OrderPayments").
		Preload("OrderPayments.Cashier").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByIDWithCategoryPreload(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("OrderItems").
		Preload("OrderItems.ProductVariant").
		Preload("OrderItems.ProductVariant.Product").
		Preload("OrderItems.ProductVariant.Product.Category").
		Preload("OrderItems.ProductVariant.Product.Category.SPKCategory").
		Preload("OrderItems.Finishings").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll(ctx context.Context, params request.PaginationParam, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string, dateBy string, categoryID *uuid.UUID, isInvoiced *bool) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Order{})

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	if len(paymentStatuses) > 0 {
		query = query.Where("payment_status IN ?", paymentStatuses)
	}

	if len(paymentMethods) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM order_payments WHERE order_payments.order_id = orders.id AND order_payments.payment_method IN ? AND order_payments.deleted_at IS NULL)", paymentMethods)
	}

	if designerID != nil && *designerID != uuid.Nil {
		query = query.Where("designer_id = ?", *designerID)
	}

	if cashierID != nil && *cashierID != uuid.Nil {
		query = query.Where("cashier_id = ?", *cashierID)
	}

	if search != "" {
		searchText := "%" + search + "%"
		query = query.Where("invoice_number ILIKE ? OR customer_name ILIKE ? OR job_number ILIKE ?", searchText, searchText, searchText)
	}

	if startDate != nil && endDate != nil {
		dateCol := "orders.created_at"
		if dateBy == "invoice" {
			dateCol = "COALESCE(orders.invoiced_at, orders.created_at)"
		}
		query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", dateCol), startDate, endDate)
	}

	switch customerType {
	case "reseller":
		query = query.Where("reseller_id IS NOT NULL")
	case "end_user":
		query = query.Where("reseller_id IS NULL")
	}

	if isInvoiced != nil {
		if *isInvoiced {
			query = query.Where("orders.invoice_number IS NOT NULL")
		} else {
			query = query.Where("orders.invoice_number IS NULL")
		}
	}

	if categoryID != nil && *categoryID != uuid.Nil {
		query = query.Where(`EXISTS (
			SELECT 1 FROM order_items oi
			JOIN product_variants pv ON pv.id = oi.product_variant_id AND pv.deleted_at IS NULL
			JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL
			WHERE oi.order_id = orders.id 
			  AND p.category_id = ? 
			  AND oi.deleted_at IS NULL
		)`, *categoryID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Preload("Reseller").
		Preload("Designer").
		Preload("Cashier").
		Preload("OrderItems").
		Preload("OrderItems.ProductVariant").
		Preload("OrderItems.ProductVariant.Product").
		Preload("OrderItems.Finishings").
		Preload("OrderPayments").
		Preload("OrderPayments.Cashier").
		Limit(params.GetLimit()).
		Offset(params.GetOffset()).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetNextJobSeq uses MAX to extract the last used sequence number, preventing
// race-condition gaps that COUNT-based approaches suffer from when cancelled
// orders are soft-deleted.
func (r *orderRepository) GetNextJobSeq(ctx context.Context, dateStr string) (int, error) {
	var maxNum *string
	prefix := "JOB/" + dateStr + "/"
	likePattern := prefix + "%"

	err := r.db.WithContext(ctx).
		Unscoped().
		Model(&domain.Order{}).
		Where("job_number LIKE ?", likePattern).
		Select("MAX(job_number)").
		Scan(&maxNum).Error
	if err != nil {
		return 0, err
	}

	if maxNum == nil {
		return 1, nil
	}

	// Extract sequence number from format "JOB/20260530/0001"
	var seq int
	_, err = fmt.Sscanf(*maxNum, prefix+"%04d", &seq)
	if err != nil {
		return 1, nil
	}

	return seq + 1, nil
}

// GetNextInvSeq uses MAX to extract the last used invoice sequence number.
func (r *orderRepository) GetNextInvSeq(ctx context.Context, dateStr string) (int, error) {
	var maxNum *string
	prefix := "INV/" + dateStr + "/"
	likePattern := prefix + "%"

	err := r.db.WithContext(ctx).
		Unscoped().
		Model(&domain.Order{}).
		Where("invoice_number LIKE ?", likePattern).
		Select("MAX(invoice_number)").
		Scan(&maxNum).Error
	if err != nil {
		return 0, err
	}

	if maxNum == nil {
		return 1, nil
	}

	var seq int
	_, err = fmt.Sscanf(*maxNum, prefix+"%05d", &seq)
	if err != nil {
		return 1, nil
	}

	return seq + 1, nil
}

func (r *orderRepository) FindFinishingsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Finishing, error) {
	var finishings []domain.Finishing
	if len(ids) == 0 {
		return finishings, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&finishings).Error
	if err != nil {
		return nil, err
	}
	return finishings, nil
}

func (r *orderRepository) CreateFinishing(ctx context.Context, finishing *domain.Finishing) error {
	return r.db.WithContext(ctx).Create(finishing).Error
}

func (r *orderRepository) FindAllFinishings(ctx context.Context) ([]domain.Finishing, error) {
	var finishings []domain.Finishing
	err := r.db.WithContext(ctx).Order("name ASC").Find(&finishings).Error
	if err != nil {
		return nil, err
	}
	return finishings, nil
}

func (r *orderRepository) CreatePayment(ctx context.Context, payment *domain.OrderPayment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *orderRepository) ReplaceItems(ctx context.Context, orderID uuid.UUID, items []domain.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var oldItems []domain.OrderItem
		if err := tx.Where("order_id = ?", orderID).Find(&oldItems).Error; err != nil {
			return err
		}
		for _, item := range oldItems {
			if err := tx.Model(&item).Association("Finishings").Clear(); err != nil {
				return err
			}
		}
		if err := tx.Where("order_id = ?", orderID).Delete(&domain.OrderItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = orderID
			// Set new ID to ensure it is inserted as a fresh record
			items[i].ID = uuid.New()
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) GetReportsWidgets(ctx context.Context, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string, dateBy string) (*domain.OrderReportsWidgetsRes, error) {
	query := r.db.WithContext(ctx).Model(&domain.Order{})

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	} else {
		// Default reports exclude DRAFT, CANCELLED, and REFUND
		query = query.Where("status NOT IN ?", []string{domain.StatusDraft, domain.StatusCancelled, domain.StatusRefund})
	}

	if len(paymentStatuses) > 0 {
		query = query.Where("payment_status IN ?", paymentStatuses)
	}

	if len(paymentMethods) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM order_payments WHERE order_payments.order_id = orders.id AND order_payments.payment_method IN ? AND order_payments.deleted_at IS NULL)", paymentMethods)
	}

	if designerID != nil && *designerID != uuid.Nil {
		query = query.Where("designer_id = ?", *designerID)
	}

	if cashierID != nil && *cashierID != uuid.Nil {
		query = query.Where("cashier_id = ?", *cashierID)
	}

	if search != "" {
		searchText := "%" + search + "%"
		query = query.Where("invoice_number ILIKE ? OR customer_name ILIKE ? OR job_number ILIKE ?", searchText, searchText, searchText)
	}

	if startDate != nil && endDate != nil {
		dateCol := "orders.created_at"
		if dateBy == "invoice" {
			dateCol = "COALESCE(orders.invoiced_at, orders.created_at)"
		}
		query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", dateCol), startDate, endDate)
	}

	switch customerType {
	case "reseller":
		query = query.Where("reseller_id IS NOT NULL")
	case "end_user":
		query = query.Where("reseller_id IS NULL")
	}

	// 1. Get OmsetPenjualan (SUM of grand_total)
	var omset float64
	err := query.Select("COALESCE(SUM(grand_total), 0)").Scan(&omset).Error
	if err != nil {
		return nil, err
	}

	// 2. Get TotalPiutang (SUM of grand_total - amount_paid for unpaid & partial paid)
	var totalPiutang float64
	err = query.Session(&gorm.Session{}).
		Where("payment_status IN ?", []string{domain.PaymentStatusUnpaid, domain.PaymentStatusPartialPaid}).
		Select("COALESCE(SUM(grand_total - amount_paid), 0)").
		Scan(&totalPiutang).Error
	if err != nil {
		return nil, err
	}

	// 3. Get BelumLunasCount (COUNT of unpaid & partial paid invoices)
	var belumLunasCount int64
	err = query.Session(&gorm.Session{}).
		Where("payment_status IN ?", []string{domain.PaymentStatusUnpaid, domain.PaymentStatusPartialPaid}).
		Count(&belumLunasCount).Error
	if err != nil {
		return nil, err
	}

	return &domain.OrderReportsWidgetsRes{
		OmsetPenjualan:  omset,
		TotalPiutang:    totalPiutang,
		BelumLunasCount: belumLunasCount,
	}, nil
}

func (r *orderRepository) GetSalesReportWidgets(ctx context.Context, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string, dateBy string, categoryID *uuid.UUID) (*domain.SalesReportWidgetsRes, error) {
	query := r.db.WithContext(ctx).Model(&domain.Order{})

	// Sales reports only count invoiced orders
	query = query.Where("orders.invoice_number IS NOT NULL")

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	} else {
		// Default reports exclude DRAFT, CANCELLED, and REFUND
		query = query.Where("status NOT IN ?", []string{domain.StatusDraft, domain.StatusCancelled, domain.StatusRefund})
	}

	if len(paymentStatuses) > 0 {
		query = query.Where("payment_status IN ?", paymentStatuses)
	}

	if len(paymentMethods) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM order_payments WHERE order_payments.order_id = orders.id AND order_payments.payment_method IN ? AND order_payments.deleted_at IS NULL)", paymentMethods)
	}

	if designerID != nil && *designerID != uuid.Nil {
		query = query.Where("designer_id = ?", *designerID)
	}

	if cashierID != nil && *cashierID != uuid.Nil {
		query = query.Where("cashier_id = ?", *cashierID)
	}

	if search != "" {
		searchText := "%" + search + "%"
		query = query.Where("invoice_number ILIKE ? OR customer_name ILIKE ? OR job_number ILIKE ?", searchText, searchText, searchText)
	}

	if startDate != nil && endDate != nil {
		dateCol := "orders.created_at"
		if dateBy == "invoice" {
			dateCol = "COALESCE(orders.invoiced_at, orders.created_at)"
		}
		query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", dateCol), startDate, endDate)
	}

	switch customerType {
	case "reseller":
		query = query.Where("reseller_id IS NOT NULL")
	case "end_user":
		query = query.Where("reseller_id IS NULL")
	}

	if categoryID != nil && *categoryID != uuid.Nil {
		query = query.Where(`EXISTS (
			SELECT 1 FROM order_items oi
			JOIN product_variants pv ON pv.id = oi.product_variant_id AND pv.deleted_at IS NULL
			JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL
			WHERE oi.order_id = orders.id 
			  AND p.category_id = ? 
			  AND oi.deleted_at IS NULL
		)`, *categoryID)
	}

	// 1. Get OmsetPenjualan
	var omset float64
	if categoryID != nil && *categoryID != uuid.Nil {
		err := query.Session(&gorm.Session{}).
			Joins("JOIN order_items oi ON oi.order_id = orders.id AND oi.deleted_at IS NULL").
			Joins("JOIN product_variants pv ON pv.id = oi.product_variant_id AND pv.deleted_at IS NULL").
			Joins("JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL").
			Where("p.category_id = ?", *categoryID).
			Select("COALESCE(SUM(oi.subtotal), 0)").
			Scan(&omset).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := query.Session(&gorm.Session{}).Select("COALESCE(SUM(grand_total), 0)").Scan(&omset).Error
		if err != nil {
			return nil, err
		}
	}

	// 2. Get VolumeTransaksi (COUNT of orders)
	var volumeTransaksi int64
	err := query.Session(&gorm.Session{}).Count(&volumeTransaksi).Error
	if err != nil {
		return nil, err
	}

	// 3. Get TotalProdukTerjual (SUM of order_items.quantity)
	var totalProdukTerjual int64
	qtyQuery := query.Session(&gorm.Session{}).
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("order_items.deleted_at IS NULL")

	if categoryID != nil && *categoryID != uuid.Nil {
		qtyQuery = qtyQuery.
			Joins("JOIN product_variants pv ON pv.id = order_items.product_variant_id AND pv.deleted_at IS NULL").
			Joins("JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL").
			Where("p.category_id = ?", *categoryID)
	}

	err = qtyQuery.Select("COALESCE(SUM(order_items.quantity), 0)").Scan(&totalProdukTerjual).Error
	if err != nil {
		return nil, err
	}

	// 4. Get LunasCount (COUNT of paid orders)
	var lunasCount int64
	err = query.Session(&gorm.Session{}).
		Where("payment_status = ?", domain.PaymentStatusPaid).
		Count(&lunasCount).Error
	if err != nil {
		return nil, err
	}

	// 5. Get BelumLunasCount (COUNT of unpaid & partial paid orders)
	var belumLunasCount int64
	err = query.Session(&gorm.Session{}).
		Where("payment_status IN ?", []string{domain.PaymentStatusUnpaid, domain.PaymentStatusPartialPaid}).
		Count(&belumLunasCount).Error
	if err != nil {
		return nil, err
	}

	return &domain.SalesReportWidgetsRes{
		OmsetPenjualan:     omset,
		VolumeTransaksi:    volumeTransaksi,
		TotalProdukTerjual: totalProdukTerjual,
		LunasCount:         lunasCount,
		BelumLunasCount:    belumLunasCount,
	}, nil
}

func (r *orderRepository) GetSalesTrend(ctx context.Context, trendType string, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string, dateBy string, categoryID *uuid.UUID) ([]domain.SalesTrendItem, error) {
	subQuery := r.db.WithContext(ctx).Model(&domain.Order{})

	// Sales reports only count invoiced orders
	subQuery = subQuery.Where("orders.invoice_number IS NOT NULL")

	if len(statuses) > 0 {
		subQuery = subQuery.Where("status IN ?", statuses)
	} else {
		subQuery = subQuery.Where("status NOT IN ?", []string{domain.StatusDraft, domain.StatusCancelled, domain.StatusRefund})
	}

	if len(paymentStatuses) > 0 {
		subQuery = subQuery.Where("payment_status IN ?", paymentStatuses)
	}

	if len(paymentMethods) > 0 {
		subQuery = subQuery.Where("EXISTS (SELECT 1 FROM order_payments WHERE order_payments.order_id = orders.id AND order_payments.payment_method IN ? AND order_payments.deleted_at IS NULL)", paymentMethods)
	}

	if designerID != nil && *designerID != uuid.Nil {
		subQuery = subQuery.Where("designer_id = ?", *designerID)
	}

	if cashierID != nil && *cashierID != uuid.Nil {
		subQuery = subQuery.Where("cashier_id = ?", *cashierID)
	}

	if search != "" {
		searchText := "%" + search + "%"
		subQuery = subQuery.Where("invoice_number ILIKE ? OR customer_name ILIKE ? OR job_number ILIKE ?", searchText, searchText, searchText)
	}

	dateCol := "orders.created_at"
	dateTruncCol := "o.created_at"
	if dateBy == "invoice" {
		dateCol = "COALESCE(orders.invoiced_at, orders.created_at)"
		dateTruncCol = "COALESCE(o.invoiced_at, o.created_at)"
	}

	if startDate != nil && endDate != nil {
		subQuery = subQuery.Where(fmt.Sprintf("%s BETWEEN ? AND ?", dateCol), startDate, endDate)
	}

	switch customerType {
	case "reseller":
		subQuery = subQuery.Where("reseller_id IS NOT NULL")
	case "end_user":
		subQuery = subQuery.Where("reseller_id IS NULL")
	}

	if categoryID != nil && *categoryID != uuid.Nil {
		subQuery = subQuery.
			Joins("JOIN order_items oi ON oi.order_id = orders.id AND oi.deleted_at IS NULL").
			Joins("JOIN product_variants pv ON pv.id = oi.product_variant_id AND pv.deleted_at IS NULL").
			Joins("JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL").
			Where("p.category_id = ?", *categoryID).
			Select("orders.id, orders.created_at, orders.invoiced_at, orders.deleted_at, SUM(oi.subtotal) AS amount").
			Group("orders.id, orders.created_at, orders.invoiced_at, orders.deleted_at")
	} else {
		subQuery = subQuery.Select("orders.id, orders.created_at, orders.invoiced_at, orders.deleted_at, orders.grand_total AS amount")
	}

	var seriesSelect string
	var groupByExpr string
	var labelExpr string
	var truncField string

	switch trendType {
	case "weekly":
		truncField = "week"
		seriesSelect = "generate_series(date_trunc('week', NOW() - INTERVAL '5 weeks'), date_trunc('week', NOW()), '1 week'::interval) AS period_start"
		groupByExpr = "d.period_start"
		labelExpr = "to_char(d.period_start, 'YYYY-MM-DD')"
	case "monthly":
		truncField = "month"
		seriesSelect = "generate_series(date_trunc('month', NOW() - INTERVAL '11 months'), date_trunc('month', NOW()), '1 month'::interval) AS period_start"
		groupByExpr = "d.period_start"
		labelExpr = "to_char(d.period_start, 'Mon YYYY')"
	case "yearly":
		truncField = "year"
		seriesSelect = "generate_series(date_trunc('year', NOW() - INTERVAL '4 years'), date_trunc('year', NOW()), '1 year'::interval) AS period_start"
		groupByExpr = "d.period_start"
		labelExpr = "to_char(d.period_start, 'YYYY')"
	default:
		return nil, fmt.Errorf("invalid trend type: %s", trendType)
	}

	var results []domain.SalesTrendItem
	err := r.db.WithContext(ctx).
		Table(fmt.Sprintf("(SELECT %s) d", seriesSelect)).
		Joins(fmt.Sprintf("LEFT JOIN (?) o ON date_trunc('%s', %s) = d.period_start AND o.deleted_at IS NULL", truncField, dateTruncCol), subQuery).
		Select(fmt.Sprintf("%s AS label, COALESCE(SUM(o.amount), 0) AS total", labelExpr)).
		Group(groupByExpr).
		Order("d.period_start ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *orderRepository) GetCategorySales(ctx context.Context, startDate *time.Time, endDate *time.Time, dateBy string) ([]domain.CategorySalesItem, error) {
	query := r.db.WithContext(ctx).Table("categories c").
		Select("c.id AS category_id, c.name AS category_name, COALESCE(SUM(oi.subtotal), 0) AS total_sales").
		Joins("LEFT JOIN products p ON p.category_id = c.id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN product_variants pv ON pv.product_id = p.id AND pv.deleted_at IS NULL").
		Joins("LEFT JOIN order_items oi ON oi.product_variant_id = pv.id AND oi.deleted_at IS NULL")

	dateCol := "o.created_at"
	if dateBy == "invoice" {
		dateCol = "COALESCE(o.invoiced_at, o.created_at)"
	}

	if startDate != nil && endDate != nil {
		query = query.Joins(fmt.Sprintf("LEFT JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL AND o.status NOT IN ('DRAFT', 'CANCELLED', 'REFUND') AND o.invoice_number IS NOT NULL AND %s BETWEEN ? AND ?", dateCol), startDate, endDate)
	} else {
		query = query.Joins("LEFT JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL AND o.status NOT IN ('DRAFT', 'CANCELLED', 'REFUND') AND o.invoice_number IS NOT NULL")
	}

	var results []domain.CategorySalesItem
	err := query.Where("c.deleted_at IS NULL").
		Group("c.id, c.name").
		Order("total_sales DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *orderRepository) GetPaymentMethodSales(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]domain.PaymentMethodSalesItem, error) {
	query := r.db.WithContext(ctx).Table("order_payments op").
		Select("op.payment_method, COALESCE(SUM(op.amount), 0) AS total_amount").
		Joins("JOIN orders o ON o.id = op.order_id AND o.deleted_at IS NULL AND o.status NOT IN ('DRAFT', 'CANCELLED', 'REFUND')")

	if startDate != nil && endDate != nil {
		query = query.Where("op.created_at BETWEEN ? AND ?", startDate, endDate)
	}

	type dbResult struct {
		PaymentMethod string
		TotalAmount   float64
	}

	var rawResults []dbResult
	err := query.Where("op.deleted_at IS NULL").
		Group("op.payment_method").
		Scan(&rawResults).Error

	if err != nil {
		return nil, err
	}

	methodsMap := map[string]float64{
		"cash":     0,
		"transfer": 0,
		"qris":     0,
	}

	for _, res := range rawResults {
		pm := strings.ToLower(res.PaymentMethod)
		if _, ok := methodsMap[pm]; ok {
			methodsMap[pm] = res.TotalAmount
		} else {
			// If there are other methods like tempo, keep them too
			methodsMap[pm] = res.TotalAmount
		}
	}

	results := []domain.PaymentMethodSalesItem{
		{PaymentMethod: "cash", TotalAmount: methodsMap["cash"]},
		{PaymentMethod: "transfer", TotalAmount: methodsMap["transfer"]},
		{PaymentMethod: "qris", TotalAmount: methodsMap["qris"]},
	}

	// Also append other methods if they exist and are non-zero (like tempo)
	for k, v := range methodsMap {
		if k != "cash" && k != "transfer" && k != "qris" && v > 0 {
			results = append(results, domain.PaymentMethodSalesItem{
				PaymentMethod: k,
				TotalAmount:   v,
			})
		}
	}

	return results, nil
}



