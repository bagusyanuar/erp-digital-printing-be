package repository

import (
	"context"
	"fmt"

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

func (r *orderRepository) FindAll(ctx context.Context, params request.PaginationParam, statuses []string, paymentStatuses []string, designerID *uuid.UUID) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Order{})

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	if len(paymentStatuses) > 0 {
		query = query.Where("payment_status IN ?", paymentStatuses)
	}

	if designerID != nil && *designerID != uuid.Nil {
		query = query.Where("designer_id = ?", *designerID)
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
