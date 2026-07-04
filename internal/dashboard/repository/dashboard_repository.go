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

func (r *dashboardRepository) GetWeeklyRevenue(ctx context.Context, targetDate time.Time) (*domain.WeeklyRevenueRes, error) {
	// Find Monday of the week
	// targetDate.Weekday() returns 0 for Sunday, 1 for Monday, etc.
	wd := int(targetDate.Weekday())
	offsetDays := wd - 1
	if wd == 0 { // Sunday should be offset by 6 days from Monday
		offsetDays = 6
	}

	monday := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location()).
		AddDate(0, 0, -offsetDays)
	sunday := monday.AddDate(0, 0, 6)
	sundayEnd := time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 999999999, sunday.Location())

	// Awal & Akhir minggu lalu
	lastMonday := monday.AddDate(0, 0, -7)
	lastSundayEnd := time.Date(monday.Year(), monday.Month(), monday.Day(), 23, 59, 59, 999999999, monday.Location()).AddDate(0, 0, -1)

	// Query aggregate revenues group by date for this week
	type dbResult struct {
		OrderDate time.Time `gorm:"column:order_date"`
		Revenue   float64   `gorm:"column:revenue"`
	}
	var dbRes []dbResult

	err := r.db.WithContext(ctx).
		Model(&orderDomain.Order{}).
		Select("DATE(created_at) AS order_date, COALESCE(SUM(grand_total), 0) AS revenue").
		Where("status IN ?", []string{orderDomain.StatusInProduction, orderDomain.StatusReadyForPickup, orderDomain.StatusCompleted}).
		Where("created_at BETWEEN ? AND ?", monday, sundayEnd).
		Group("DATE(created_at)").
		Scan(&dbRes).Error

	if err != nil {
		return nil, err
	}

	// Map DB results for quick lookup
	revenueMap := make(map[string]float64)
	var totalThisWeek float64
	for _, row := range dbRes {
		dateStr := row.OrderDate.Format("2006-01-02")
		revenueMap[dateStr] = row.Revenue
		totalThisWeek += row.Revenue
	}

	// Query total revenue for last week
	var totalLastWeek float64
	err = r.db.WithContext(ctx).
		Model(&orderDomain.Order{}).
		Select("COALESCE(SUM(grand_total), 0)").
		Where("status IN ?", []string{orderDomain.StatusInProduction, orderDomain.StatusReadyForPickup, orderDomain.StatusCompleted}).
		Where("created_at BETWEEN ? AND ?", lastMonday, lastSundayEnd).
		Scan(&totalLastWeek).Error

	if err != nil {
		return nil, err
	}

	// Hitung persentase kenaikan/penurunan
	var percentageChange float64
	if totalLastWeek > 0 {
		percentageChange = ((totalThisWeek - totalLastWeek) / totalLastWeek) * 100
	} else if totalThisWeek > 0 {
		// Jika minggu lalu 0 tapi minggu ini ada omzet, kenaikan +100%
		percentageChange = 100
	} else {
		percentageChange = 0
	}

	// Generate full 7 days response
	dayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	weeklyData := make([]domain.WeeklyRevenueItem, 7)
	for i := 0; i < 7; i++ {
		currentDay := monday.AddDate(0, 0, i)
		dateStr := currentDay.Format("2006-01-02")
		
		weeklyData[i] = domain.WeeklyRevenueItem{
			Date:    dateStr,
			DayName: dayNames[i],
			Revenue: revenueMap[dateStr],
		}
	}

	return &domain.WeeklyRevenueRes{
		TotalThisWeek:    totalThisWeek,
		TotalLastWeek:    totalLastWeek,
		PercentageChange: percentageChange,
		WeeklyData:       weeklyData,
	}, nil
}
