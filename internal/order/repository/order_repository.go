package repository

import (
	"context"
	"fmt"
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
		Preload("OrderItems.Finishings").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll(ctx context.Context, params request.PaginationParam, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string) ([]domain.Order, int64, error) {
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
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}

	switch customerType {
	case "reseller":
		query = query.Where("reseller_id IS NOT NULL")
	case "end_user":
		query = query.Where("reseller_id IS NULL")
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
	_, err = fmt.Sscanf(*maxNum, prefix+"%04d", &seq)
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

func (r *orderRepository) GetReportsWidgets(ctx context.Context, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string) (*domain.OrderReportsWidgetsRes, error) {
	query := r.db.WithContext(ctx).Model(&domain.Order{})

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	} else {
		// Default reports exclude DRAFT and CANCELLED
		query = query.Where("status NOT IN ?", []string{domain.StatusDraft, domain.StatusCancelled})
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
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
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
