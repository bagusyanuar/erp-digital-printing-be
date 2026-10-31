package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/report/domain"
	"gorm.io/gorm"
)

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) domain.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetRevenueByCategory(ctx context.Context, startDate, endDate *time.Time) ([]domain.ProfitLossItem, float64, error) {
	query := r.db.WithContext(ctx).Table("categories c").
		Select("c.id::text AS id, c.name AS name, COALESCE(SUM(oi.subtotal), 0) AS amount").
		Joins("LEFT JOIN products p ON p.category_id = c.id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN product_variants pv ON pv.product_id = p.id AND pv.deleted_at IS NULL").
		Joins("LEFT JOIN order_items oi ON oi.product_variant_id = pv.id AND oi.deleted_at IS NULL")

	if startDate != nil && endDate != nil {
		query = query.Joins("LEFT JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL AND o.status NOT IN ('DRAFT', 'CANCELLED', 'REFUND') AND o.invoice_number IS NOT NULL AND o.created_at BETWEEN ? AND ?", startDate, endDate)
	} else {
		query = query.Joins("LEFT JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL AND o.status NOT IN ('DRAFT', 'CANCELLED', 'REFUND') AND o.invoice_number IS NOT NULL")
	}

	var rawItems []struct {
		ID     string  `gorm:"column:id"`
		Name   string  `gorm:"column:name"`
		Amount float64 `gorm:"column:amount"`
	}

	err := query.Where("c.deleted_at IS NULL").
		Group("c.id, c.name").
		Having("COALESCE(SUM(oi.subtotal), 0) > 0").
		Order("amount DESC").
		Scan(&rawItems).Error
	if err != nil {
		return nil, 0, err
	}

	var total float64
	items := make([]domain.ProfitLossItem, len(rawItems))
	for i, it := range rawItems {
		total += it.Amount
		name := it.Name
		if !strings.HasPrefix(strings.ToLower(name), "penjualan") {
			name = "Penjualan " + name
		}
		items[i] = domain.ProfitLossItem{
			ID:     it.ID,
			Name:   name,
			Amount: it.Amount,
		}
	}

	return items, total, nil
}

func (r *reportRepository) GetCOGSByCategory(ctx context.Context, startDate, endDate *time.Time) ([]domain.ProfitLossItem, float64, error) {
	query := r.db.WithContext(ctx).Table("expense_items ei").
		Select("c.id::text AS id, c.name AS name, COALESCE(SUM(ei.amount), 0) AS amount").
		Joins("JOIN expenses e ON e.id = ei.expense_id AND e.deleted_at IS NULL").
		Joins("JOIN expense_categories ec ON ec.id = ei.expense_category_id AND ec.deleted_at IS NULL").
		Joins("JOIN categories c ON c.id = ec.product_category_id AND c.deleted_at IS NULL").
		Where("ei.deleted_at IS NULL").
		Where("ec.group = 'PRODUCTION'").
		Where("ec.product_category_id IS NOT NULL").
		Where("ec.id != '00000000-0000-0000-0000-000000000000'")

	if startDate != nil {
		query = query.Where("e.expense_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("e.expense_date <= ?", *endDate)
	}

	var rawItems []struct {
		ID     string  `gorm:"column:id"`
		Name   string  `gorm:"column:name"`
		Amount float64 `gorm:"column:amount"`
	}

	err := query.Group("c.id, c.name").
		Having("COALESCE(SUM(ei.amount), 0) > 0").
		Order("amount DESC").
		Scan(&rawItems).Error
	if err != nil {
		return nil, 0, err
	}

	var total float64
	items := make([]domain.ProfitLossItem, len(rawItems))
	for i, it := range rawItems {
		total += it.Amount
		name := it.Name
		if !strings.HasPrefix(strings.ToLower(name), "bahan") {
			name = "Bahan Baku " + name
		}
		items[i] = domain.ProfitLossItem{
			ID:     it.ID,
			Name:   name,
			Amount: it.Amount,
		}
	}

	return items, total, nil
}

func (r *reportRepository) GetOperationalExpenses(ctx context.Context, startDate, endDate *time.Time) ([]domain.ProfitLossItem, float64, error) {
	query := r.db.WithContext(ctx).Table("expense_items ei").
		Select("ec.id::text AS id, ec.name AS name, COALESCE(SUM(ei.amount), 0) AS amount").
		Joins("JOIN expenses e ON e.id = ei.expense_id AND e.deleted_at IS NULL").
		Joins("JOIN expense_categories ec ON ec.id = ei.expense_category_id AND ec.deleted_at IS NULL").
		Where("ei.deleted_at IS NULL").
		Where("ec.group = 'OPERATIONAL'").
		Where("ec.id != '00000000-0000-0000-0000-000000000000'")

	if startDate != nil {
		query = query.Where("e.expense_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("e.expense_date <= ?", *endDate)
	}

	var rawItems []struct {
		ID     string  `gorm:"column:id"`
		Name   string  `gorm:"column:name"`
		Amount float64 `gorm:"column:amount"`
	}

	err := query.Group("ec.id, ec.name").
		Having("COALESCE(SUM(ei.amount), 0) > 0").
		Order("amount DESC").
		Scan(&rawItems).Error
	if err != nil {
		return nil, 0, err
	}

	var total float64
	items := make([]domain.ProfitLossItem, len(rawItems))
	for i, it := range rawItems {
		total += it.Amount
		items[i] = domain.ProfitLossItem{
			ID:     it.ID,
			Name:   it.Name,
			Amount: it.Amount,
		}
	}

	return items, total, nil
}

func (r *reportRepository) GetMonthlyProfitLossTrend(ctx context.Context, months int) ([]domain.ProfitLossTrendItem, error) {
	if months <= 0 {
		months = 6
	} else if months > 24 {
		months = 24
	}

	intervalMonths := months - 1

	sqlQuery := fmt.Sprintf(`
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', NOW() - INTERVAL '%d months'),
				date_trunc('month', NOW()),
				'1 month'::interval
			) AS month_date
		),
		monthly_rev AS (
			SELECT 
				date_trunc('month', o.created_at) AS month_date,
				COALESCE(SUM(oi.subtotal), 0) AS total_rev
			FROM orders o
			JOIN order_items oi ON oi.order_id = o.id AND oi.deleted_at IS NULL
			WHERE o.deleted_at IS NULL 
			  AND o.status NOT IN ('DRAFT', 'CANCELLED', 'REFUND')
			  AND o.invoice_number IS NOT NULL
			  AND o.created_at >= date_trunc('month', NOW() - INTERVAL '%d months')
			GROUP BY 1
		),
		monthly_cogs AS (
			SELECT 
				date_trunc('month', e.expense_date) AS month_date,
				COALESCE(SUM(ei.amount), 0) AS total_cogs
			FROM expenses e
			JOIN expense_items ei ON ei.expense_id = e.id AND ei.deleted_at IS NULL
			JOIN expense_categories ec ON ec.id = ei.expense_category_id AND ec.deleted_at IS NULL
			WHERE e.deleted_at IS NULL
			  AND ec.group = 'PRODUCTION'
			  AND ec.id != '00000000-0000-0000-0000-000000000000'
			  AND e.expense_date >= date_trunc('month', NOW() - INTERVAL '%d months')
			GROUP BY 1
		),
		monthly_opex AS (
			SELECT 
				date_trunc('month', e.expense_date) AS month_date,
				COALESCE(SUM(ei.amount), 0) AS total_opex
			FROM expenses e
			JOIN expense_items ei ON ei.expense_id = e.id AND ei.deleted_at IS NULL
			JOIN expense_categories ec ON ec.id = ei.expense_category_id AND ec.deleted_at IS NULL
			WHERE e.deleted_at IS NULL
			  AND ec.group = 'OPERATIONAL'
			  AND ec.id != '00000000-0000-0000-0000-000000000000'
			  AND e.expense_date >= date_trunc('month', NOW() - INTERVAL '%d months')
			GROUP BY 1
		)
		SELECT 
			m.month_date,
			COALESCE(r.total_rev, 0) AS pendapatan,
			COALESCE(c.total_cogs, 0) AS hpp,
			COALESCE(op.total_opex, 0) AS opex,
			(COALESCE(r.total_rev, 0) - COALESCE(c.total_cogs, 0) - COALESCE(op.total_opex, 0)) AS laba_bersih
		FROM months m
		LEFT JOIN monthly_rev r ON r.month_date = m.month_date
		LEFT JOIN monthly_cogs c ON c.month_date = m.month_date
		LEFT JOIN monthly_opex op ON op.month_date = m.month_date
		ORDER BY m.month_date ASC
	`, intervalMonths, intervalMonths, intervalMonths, intervalMonths)

	type rawTrend struct {
		MonthDate  time.Time `gorm:"column:month_date"`
		Pendapatan float64   `gorm:"column:pendapatan"`
		HPP        float64   `gorm:"column:hpp"`
		Opex       float64   `gorm:"column:opex"`
		LabaBersih float64   `gorm:"column:laba_bersih"`
	}

	var rawResults []rawTrend
	if err := r.db.WithContext(ctx).Raw(sqlQuery).Scan(&rawResults).Error; err != nil {
		return nil, err
	}

	indoMonths := [...]string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}

	results := make([]domain.ProfitLossTrendItem, len(rawResults))
	for i, row := range rawResults {
		mNum := int(row.MonthDate.Month())
		label := fmt.Sprintf("%s %d", indoMonths[mNum], row.MonthDate.Year())
		results[i] = domain.ProfitLossTrendItem{
			Month:      label,
			PeriodDate: row.MonthDate.Format("2006-01-02"),
			Pendapatan: row.Pendapatan,
			HPP:        row.HPP,
			Opex:       row.Opex,
			LabaBersih: row.LabaBersih,
		}
	}

	return results, nil
}
