package repository

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/dashboard/domain"
	orderDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	expenseDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) domain.DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetWidgets(ctx context.Context, startDate *time.Time, endDate *time.Time) (*domain.DashboardWidgetsRes, error) {
	var res domain.DashboardWidgetsRes

	// 1. Total Omzet & Volume Transaksi
	ordersQuery := r.db.WithContext(ctx).Model(&orderDomain.Order{}).
		Where("status NOT IN ?", []string{orderDomain.StatusDraft, orderDomain.StatusCancelled, orderDomain.StatusRefund})

	if startDate != nil {
		ordersQuery = ordersQuery.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		ordersQuery = ordersQuery.Where("created_at <= ?", *endDate)
	}

	var omzet float64
	if err := ordersQuery.Select("COALESCE(SUM(grand_total), 0)").Scan(&omzet).Error; err != nil {
		return nil, err
	}
	res.TotalOmzet = omzet

	var volume int64
	if err := ordersQuery.Count(&volume).Error; err != nil {
		return nil, err
	}
	res.VolumeTransaksi = volume

	// 2. Total Pendapatan
	paymentsQuery := r.db.WithContext(ctx).Model(&orderDomain.OrderPayment{})
	if startDate != nil {
		paymentsQuery = paymentsQuery.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		paymentsQuery = paymentsQuery.Where("created_at <= ?", *endDate)
	}

	var pendapatan float64
	if err := paymentsQuery.Select("COALESCE(SUM(amount), 0)").Scan(&pendapatan).Error; err != nil {
		return nil, err
	}
	res.TotalPendapatan = pendapatan

	// 3. Total Pengeluaran
	expensesQuery := r.db.WithContext(ctx).Model(&expenseDomain.Expense{})
	if startDate != nil {
		expensesQuery = expensesQuery.Where("expense_date >= ?", *startDate)
	}
	if endDate != nil {
		expensesQuery = expensesQuery.Where("expense_date <= ?", *endDate)
	}

	var pengeluaran float64
	if err := expensesQuery.Select("COALESCE(SUM(amount), 0)").Scan(&pengeluaran).Error; err != nil {
		return nil, err
	}
	res.TotalPengeluaran = pengeluaran

	return &res, nil
}
