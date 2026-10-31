package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	orderDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	reportDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/report/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type reportUsecase struct {
	orderRepo  orderDomain.OrderRepository
	reportRepo reportDomain.ReportRepository
	logger     *zap.Logger
}

func NewReportUsecase(orderRepo orderDomain.OrderRepository, reportRepo reportDomain.ReportRepository, logger *zap.Logger) reportDomain.ReportUsecase {
	return &reportUsecase{
		orderRepo:  orderRepo,
		reportRepo: reportRepo,
		logger:     logger,
	}
}

func (u *reportUsecase) ExportSalesReport(ctx context.Context, startDate *time.Time, endDate *time.Time) (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			u.logger.Error("failed to close excel file", zap.Error(err))
		}
	}()

	params := request.PaginationParam{Limit: 100000, Page: 1}
	isInvoiced := true
	orders, _, err := u.orderRepo.FindAll(ctx, params, nil, nil, nil, nil, nil, "", startDate, endDate, "", "invoice", nil, &isInvoiced)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	f.SetSheetName("Sheet1", "Ringkasan")
	f.SetCellValue("Ringkasan", "A1", "Ringkasan Laporan Penjualan")
	if startDate != nil && endDate != nil {
		f.SetCellValue("Ringkasan", "A2", fmt.Sprintf("Periode: %s s/d %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
	}

	f.SetCellValue("Ringkasan", "A4", "Total Omzet")
	f.SetCellValue("Ringkasan", "B4", "Total Order")

	var totalOmzet float64
	for _, order := range orders {
		totalOmzet += order.GrandTotal
	}
	f.SetCellValue("Ringkasan", "A5", totalOmzet)
	f.SetCellValue("Ringkasan", "B5", len(orders))

	sheetName := "Detail Transaksi"
	f.NewSheet(sheetName)

	headers := []string{
		"No. Invoice", "No. SPK", "Tanggal", "Pelanggan", "No. Telepon",
		"Kasir", "Status Pembayaran", "Grand Total", "Total Terbayar",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	for i, order := range orders {
		row := i + 2
		inv := "-"
		if order.InvoiceNumber != nil {
			inv = *order.InvoiceNumber
		}
		custName := "-"
		if order.CustomerName != nil {
			custName = *order.CustomerName
		}
		custPhone := "-"
		if order.CustomerPhone != nil {
			custPhone = *order.CustomerPhone
		}
		cashierName := "-"
		if order.Cashier != nil {
			cashierName = order.Cashier.Username
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), inv)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), order.JobNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), order.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), custName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), custPhone)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), cashierName)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), order.PaymentStatus)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), order.GrandTotal)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), order.AmountPaid)
	}

	return f, nil
}

func (u *reportUsecase) GetProfitLoss(ctx context.Context, startDate, endDate *time.Time) (*reportDomain.ProfitLossRes, error) {
	now := time.Now()
	if startDate == nil {
		s := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		startDate = &s
	}
	if endDate == nil {
		e := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, time.Local)
		endDate = &e
	}

	revItems, totalRev, err := u.reportRepo.GetRevenueByCategory(ctx, startDate, endDate)
	if err != nil {
		u.logger.Error("failed to get revenue by category", zap.Error(err))
		return nil, err
	}

	cogsItems, totalCOGS, err := u.reportRepo.GetCOGSByCategory(ctx, startDate, endDate)
	if err != nil {
		u.logger.Error("failed to get cogs by category", zap.Error(err))
		return nil, err
	}

	opexItems, totalOpex, err := u.reportRepo.GetOperationalExpenses(ctx, startDate, endDate)
	if err != nil {
		u.logger.Error("failed to get operational expenses", zap.Error(err))
		return nil, err
	}

	calcPct := func(val float64) float64 {
		if totalRev <= 0 {
			return 0
		}
		return math.Round((val/totalRev)*1000) / 10
	}

	for i := range revItems {
		revItems[i].Percentage = calcPct(revItems[i].Amount)
	}
	for i := range cogsItems {
		cogsItems[i].Percentage = calcPct(cogsItems[i].Amount)
	}
	for i := range opexItems {
		opexItems[i].Percentage = calcPct(opexItems[i].Amount)
	}

	grossProfitAmt := totalRev - totalCOGS
	var grossMargin float64
	if totalRev > 0 {
		grossMargin = math.Round((grossProfitAmt/totalRev)*1000) / 10
	}

	netProfitAmt := grossProfitAmt - totalOpex
	var netMargin float64
	if totalRev > 0 {
		netMargin = math.Round((netProfitAmt/totalRev)*1000) / 10
	}

	periodStr := fmt.Sprintf("%s s/d %s", startDate.Format("02/01/2006"), endDate.Format("02/01/2006"))

	res := &reportDomain.ProfitLossRes{
		Period:    periodStr,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Revenue: reportDomain.ProfitLossSection{
			Title: "I. Pendapatan Usaha (Revenue)",
			Items: revItems,
			Total: totalRev,
		},
		COGS: reportDomain.ProfitLossSection{
			Title: "II. Beban Pokok Penjualan (HPP / Biaya Bahan Baku)",
			Items: cogsItems,
			Total: totalCOGS,
		},
		GrossProfit: reportDomain.ProfitLossGrossProfit{
			Amount:           grossProfitAmt,
			MarginPercentage: grossMargin,
		},
		OperationalExpenses: reportDomain.ProfitLossSection{
			Title: "III. Beban Operasional (Operational Expenses)",
			Items: opexItems,
			Total: totalOpex,
		},
		NetProfit: reportDomain.ProfitLossNetProfit{
			Amount:           netProfitAmt,
			MarginPercentage: netMargin,
			IsProfit:         netProfitAmt >= 0,
		},
	}

	return res, nil
}

func (u *reportUsecase) GetProfitLossTrend(ctx context.Context, months int) ([]reportDomain.ProfitLossTrendItem, error) {
	return u.reportRepo.GetMonthlyProfitLossTrend(ctx, months)
}
