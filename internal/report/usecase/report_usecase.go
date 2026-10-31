package usecase

import (
	"context"
	"fmt"
	"time"

	orderDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	reportDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/report/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type reportUsecase struct {
	orderRepo orderDomain.OrderRepository
	logger    *zap.Logger
}

func NewReportUsecase(orderRepo orderDomain.OrderRepository, logger *zap.Logger) reportDomain.ReportUsecase {
	return &reportUsecase{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (u *reportUsecase) ExportSalesReport(ctx context.Context, startDate *time.Time, endDate *time.Time) (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			u.logger.Error("failed to close excel file", zap.Error(err))
		}
	}()

	// Ambil data order tanpa batas (atau batas sangat besar, ex: 100000)
	params := request.PaginationParam{Limit: 100000, Page: 1}
	orders, _, err := u.orderRepo.FindAll(ctx, params, nil, nil, nil, nil, nil, "", startDate, endDate, "", "invoice")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	// 1. Setup Sheet Ringkasan
	f.SetSheetName("Sheet1", "Ringkasan")
	f.SetCellValue("Ringkasan", "A1", "Ringkasan Laporan Penjualan")
	if startDate != nil && endDate != nil {
		f.SetCellValue("Ringkasan", "A2", fmt.Sprintf("Periode: %s s/d %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
	}

	// Header Ringkasan
	f.SetCellValue("Ringkasan", "A4", "Total Omzet")
	f.SetCellValue("Ringkasan", "B4", "Total Order")

	var totalOmzet float64
	for _, order := range orders {
		totalOmzet += order.GrandTotal
	}
	f.SetCellValue("Ringkasan", "A5", totalOmzet)
	f.SetCellValue("Ringkasan", "B5", len(orders))

	// 2. Setup Sheet Detail Transaksi
	f.NewSheet("Detail Transaksi")
	detailHeaders := []string{"No Invoice", "Job Number", "Tanggal", "Nama Pelanggan", "Status", "Payment Status", "Grand Total"}
	for i, header := range detailHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Detail Transaksi", cell, header)
	}

	row := 2
	for _, order := range orders {
		customerName := ""
		if order.CustomerName != nil {
			customerName = *order.CustomerName
		} else if order.Reseller != nil {
			customerName = order.Reseller.Name
		}

		inv := ""
		if order.InvoiceNumber != nil {
			inv = *order.InvoiceNumber
		}

		f.SetCellValue("Detail Transaksi", fmt.Sprintf("A%d", row), inv)
		f.SetCellValue("Detail Transaksi", fmt.Sprintf("B%d", row), order.JobNumber)
		f.SetCellValue("Detail Transaksi", fmt.Sprintf("C%d", row), order.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue("Detail Transaksi", fmt.Sprintf("D%d", row), customerName)
		f.SetCellValue("Detail Transaksi", fmt.Sprintf("E%d", row), order.Status)
		f.SetCellValue("Detail Transaksi", fmt.Sprintf("F%d", row), order.PaymentStatus)
		f.SetCellValue("Detail Transaksi", fmt.Sprintf("G%d", row), order.GrandTotal)
		row++
	}

	// 3. Setup Sheet Detail Produk Terjual
	f.NewSheet("Detail Produk Terjual")
	productHeaders := []string{"Job Number", "Nama Produk", "UOM", "Ukuran", "Qty", "Harga Satuan", "Subtotal"}
	for i, header := range productHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Detail Produk Terjual", cell, header)
	}

	pRow := 2
	for _, order := range orders {
		for _, item := range order.OrderItems {
			productName := ""
			if item.ProductVariant != nil {
				if item.ProductVariant.Product.ID != uuid.Nil {
					productName = item.ProductVariant.Product.Name + " - " + item.ProductVariant.VariantName
				} else {
					productName = item.ProductVariant.VariantName
				}
			}

			ukuran := ""
			if item.LengthM != nil && item.WidthM != nil {
				ukuran = fmt.Sprintf("%.2fx%.2f m", *item.LengthM, *item.WidthM)
			}

			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("A%d", pRow), order.JobNumber)
			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("B%d", pRow), productName)
			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("C%d", pRow), item.UOM)
			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("D%d", pRow), ukuran)
			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("E%d", pRow), item.Quantity)
			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("F%d", pRow), item.PricePerUnit)
			f.SetCellValue("Detail Produk Terjual", fmt.Sprintf("G%d", pRow), item.Subtotal)
			pRow++
		}
	}

	return f, nil
}
