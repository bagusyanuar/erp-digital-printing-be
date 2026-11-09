package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"
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

func mapOrderStatus(status string) string {
	switch strings.ToUpper(status) {
	case orderDomain.StatusCancelled:
		return "Dibatalkan"
	case orderDomain.StatusRefund:
		return "Refund"
	default:
		return "Selesai"
	}
}

func mapPaymentStatus(paymentStatus string, grandTotal, amountPaid float64, hasRefund bool, isRefund bool) string {
	isPaid := paymentStatus == orderDomain.PaymentStatusPaid || (grandTotal > 0 && amountPaid >= grandTotal)
	statusStr := "Belum Lunas"
	if isPaid {
		statusStr = "Lunas"
	}
	if hasRefund && !isRefund {
		statusStr = fmt.Sprintf("%s (Refund Sebagian)", statusStr)
	}
	return statusStr
}

func (u *reportUsecase) ExportSalesReport(ctx context.Context, startDate *time.Time, endDate *time.Time) (*excelize.File, error) {
	params := request.PaginationParam{Limit: 100000, Page: 1}
	isInvoiced := true
	orders, _, err := u.orderRepo.FindAll(ctx, params, nil, nil, nil, nil, nil, "", startDate, endDate, "", "invoice", nil, &isInvoiced)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	f := excelize.NewFile()

	sheet1 := "Penjualan"
	sheet2 := "Detail Penjualan"

	_ = f.SetSheetName("Sheet1", sheet1)
	_, err = f.NewSheet(sheet2)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(0)

	var periodeText string
	if startDate != nil && endDate != nil {
		if startDate.Format("2006-01-02") == endDate.Format("2006-01-02") {
			periodeText = fmt.Sprintf("Periode: %s", startDate.Format("02/01/2006"))
		} else {
			periodeText = fmt.Sprintf("Periode: %s s/d %s", startDate.Format("02/01/2006"), endDate.Format("02/01/2006"))
		}
	} else if startDate != nil {
		periodeText = fmt.Sprintf("Periode: Mulai %s", startDate.Format("02/01/2006"))
	} else if endDate != nil {
		periodeText = fmt.Sprintf("Periode: Sampai %s", endDate.Format("02/01/2006"))
	} else {
		periodeText = "Periode: Semua Periode"
	}

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1E3A8A", Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	periodStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: false, Color: "4B5563", Size: 10, Italic: true},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E3A8A"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "D1D5DB", Style: 1},
			{Type: "bottom", Color: "D1D5DB", Style: 1},
			{Type: "left", Color: "D1D5DB", Style: 1},
			{Type: "right", Color: "D1D5DB", Style: 1},
		},
	})

	textLeftStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	textCenterStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	numFmt := "#,##0"
	numRightStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	totalLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1F2937"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F3F4F6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "1F2937", Style: 1},
			{Type: "bottom", Color: "1F2937", Style: 6},
			{Type: "left", Color: "1F2937", Style: 1},
			{Type: "right", Color: "1F2937", Style: 1},
		},
	})

	totalNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true, Color: "1F2937"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"F3F4F6"}, Pattern: 1},
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "1F2937", Style: 1},
			{Type: "bottom", Color: "1F2937", Style: 6},
			{Type: "left", Color: "1F2937", Style: 1},
			{Type: "right", Color: "1F2937", Style: 1},
		},
	})

	// Cancelled Row Styles (Soft Red)
	cancelledTextStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "991B1B"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEE2E2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FCA5A5", Style: 1},
			{Type: "bottom", Color: "FCA5A5", Style: 1},
			{Type: "left", Color: "FCA5A5", Style: 1},
			{Type: "right", Color: "FCA5A5", Style: 1},
		},
	})
	cancelledCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "991B1B"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEE2E2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FCA5A5", Style: 1},
			{Type: "bottom", Color: "FCA5A5", Style: 1},
			{Type: "left", Color: "FCA5A5", Style: 1},
			{Type: "right", Color: "FCA5A5", Style: 1},
		},
	})
	cancelledNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Color: "991B1B"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"FEE2E2"}, Pattern: 1},
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FCA5A5", Style: 1},
			{Type: "bottom", Color: "FCA5A5", Style: 1},
			{Type: "left", Color: "FCA5A5", Style: 1},
			{Type: "right", Color: "FCA5A5", Style: 1},
		},
	})

	// Refund Row Styles (Soft Purple)
	refundTextStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "5B21B6"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"EDE9FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "DDD6FE", Style: 1},
			{Type: "bottom", Color: "DDD6FE", Style: 1},
			{Type: "left", Color: "DDD6FE", Style: 1},
			{Type: "right", Color: "DDD6FE", Style: 1},
		},
	})
	refundCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "5B21B6"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"EDE9FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "DDD6FE", Style: 1},
			{Type: "bottom", Color: "DDD6FE", Style: 1},
			{Type: "left", Color: "DDD6FE", Style: 1},
			{Type: "right", Color: "DDD6FE", Style: 1},
		},
	})
	refundNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Color: "5B21B6"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"EDE9FE"}, Pattern: 1},
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "DDD6FE", Style: 1},
			{Type: "bottom", Color: "DDD6FE", Style: 1},
			{Type: "left", Color: "DDD6FE", Style: 1},
			{Type: "right", Color: "DDD6FE", Style: 1},
		},
	})

	// Partial Refund Row Styles (Soft Amber)
	partialRefundTextStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "92400E"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF3C7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FDE68A", Style: 1},
			{Type: "bottom", Color: "FDE68A", Style: 1},
			{Type: "left", Color: "FDE68A", Style: 1},
			{Type: "right", Color: "FDE68A", Style: 1},
		},
	})
	partialRefundCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "92400E"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF3C7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FDE68A", Style: 1},
			{Type: "bottom", Color: "FDE68A", Style: 1},
			{Type: "left", Color: "FDE68A", Style: 1},
			{Type: "right", Color: "FDE68A", Style: 1},
		},
	})
	partialRefundNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Color: "92400E"},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"FEF3C7"}, Pattern: 1},
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "FDE68A", Style: 1},
			{Type: "bottom", Color: "FDE68A", Style: 1},
			{Type: "left", Color: "FDE68A", Style: 1},
			{Type: "right", Color: "FDE68A", Style: 1},
		},
	})

	// Sheet 1: Header Judul & Periode
	_ = f.SetRowHeight(sheet1, 1, 24)
	_ = f.SetRowHeight(sheet1, 2, 18)
	_ = f.SetRowHeight(sheet1, 4, 26)

	_ = f.SetCellValue(sheet1, "A1", "LAPORAN PENJUALAN")
	_ = f.SetCellStyle(sheet1, "A1", "A1", titleStyle)
	_ = f.SetCellValue(sheet1, "A2", periodeText)
	_ = f.SetCellStyle(sheet1, "A2", "A2", periodStyle)

	sheet1Headers := []string{
		"No.", "No. Invoice", "No. SPK", "Tanggal", "Pelanggan",
		"No. Telepon", "Kasir", "Status Transaksi", "Status Bayar", "Metode Bayar",
		"Grand Total", "Terbayar", "Sisa Piutang",
	}

	for i, h := range sheet1Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet1, cell, h)
		_ = f.SetCellStyle(sheet1, cell, cell, headerStyle)
	}

	_ = f.SetColWidth(sheet1, "A", "A", 6)
	_ = f.SetColWidth(sheet1, "B", "C", 20)
	_ = f.SetColWidth(sheet1, "D", "D", 18)
	_ = f.SetColWidth(sheet1, "E", "E", 25)
	_ = f.SetColWidth(sheet1, "F", "G", 18)
	_ = f.SetColWidth(sheet1, "H", "J", 18)
	_ = f.SetColWidth(sheet1, "K", "M", 18)

	// Sheet 2: Header Judul & Periode
	_ = f.SetRowHeight(sheet2, 1, 24)
	_ = f.SetRowHeight(sheet2, 2, 18)
	_ = f.SetRowHeight(sheet2, 4, 26)

	_ = f.SetCellValue(sheet2, "A1", "DETAIL PENJUALAN")
	_ = f.SetCellStyle(sheet2, "A1", "A1", titleStyle)
	_ = f.SetCellValue(sheet2, "A2", periodeText)
	_ = f.SetCellStyle(sheet2, "A2", "A2", periodStyle)

	sheet2Headers := []string{
		"No.", "No. Invoice", "No. SPK", "Kategori", "Nama Produk",
		"Varian", "Ukuran", "Qty", "Satuan", "Harga Satuan",
		"Catatan", "Subtotal",
	}

	for i, h := range sheet2Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet2, cell, h)
		_ = f.SetCellStyle(sheet2, cell, cell, headerStyle)
	}

	_ = f.SetColWidth(sheet2, "A", "A", 6)
	_ = f.SetColWidth(sheet2, "B", "C", 20)
	_ = f.SetColWidth(sheet2, "D", "D", 18)
	_ = f.SetColWidth(sheet2, "E", "F", 25)
	_ = f.SetColWidth(sheet2, "G", "G", 16)
	_ = f.SetColWidth(sheet2, "H", "I", 10)
	_ = f.SetColWidth(sheet2, "J", "J", 16)
	_ = f.SetColWidth(sheet2, "K", "K", 25)
	_ = f.SetColWidth(sheet2, "L", "L", 18)

	itemRowIdx := 5
	itemSeq := 1

	for i, order := range orders {
		orderRowIdx := i + 5

		inv := "-"
		if order.InvoiceNumber != nil && *order.InvoiceNumber != "" {
			inv = *order.InvoiceNumber
		}
		custName := "-"
		if order.CustomerName != nil && *order.CustomerName != "" {
			custName = *order.CustomerName
		} else if order.Reseller != nil {
			custName = order.Reseller.Name + " (Reseller)"
		}
		custPhone := "-"
		if order.CustomerPhone != nil && *order.CustomerPhone != "" {
			custPhone = *order.CustomerPhone
		}
		cashierName := "-"
		if order.Cashier != nil {
			cashierName = order.Cashier.Username
		}

		hasRefund := false
		refundMethods := make([]string, 0)
		seenRefundMethods := make(map[string]bool)

		paymentMethods := make([]string, 0)
		seenPayMethods := make(map[string]bool)

		for _, p := range order.OrderPayments {
			if strings.ToUpper(p.PaymentType) == "REFUND" {
				hasRefund = true
				if p.PaymentMethod != "" && !seenRefundMethods[p.PaymentMethod] {
					seenRefundMethods[p.PaymentMethod] = true
					refundMethods = append(refundMethods, p.PaymentMethod)
				}
			} else {
				if p.PaymentMethod != "" && !seenPayMethods[p.PaymentMethod] {
					seenPayMethods[p.PaymentMethod] = true
					paymentMethods = append(paymentMethods, p.PaymentMethod)
				}
			}
		}

		payMethod := "-"
		if len(paymentMethods) > 0 {
			payMethod = strings.Join(paymentMethods, ", ")
			if len(refundMethods) > 0 {
				payMethod = fmt.Sprintf("%s (Refund: %s)", payMethod, strings.Join(refundMethods, ", "))
			}
		} else if len(refundMethods) > 0 {
			payMethod = fmt.Sprintf("Refund: %s", strings.Join(refundMethods, ", "))
		}

		upperStatus := strings.ToUpper(order.Status)
		isCancelled := upperStatus == orderDomain.StatusCancelled
		isRefund := upperStatus == orderDomain.StatusRefund
		isPartialRefund := hasRefund && !isRefund && !isCancelled

		sisa := order.GrandTotal - order.AmountPaid
		if isCancelled || isRefund || sisa < 0 {
			sisa = 0
		}

		var rowTextStyle, rowCenterStyle, rowNumStyle int
		if isCancelled {
			rowTextStyle = cancelledTextStyle
			rowCenterStyle = cancelledCenterStyle
			rowNumStyle = cancelledNumStyle
		} else if isRefund {
			rowTextStyle = refundTextStyle
			rowCenterStyle = refundCenterStyle
			rowNumStyle = refundNumStyle
		} else if isPartialRefund {
			rowTextStyle = partialRefundTextStyle
			rowCenterStyle = partialRefundCenterStyle
			rowNumStyle = partialRefundNumStyle
		} else {
			rowTextStyle = textLeftStyle
			rowCenterStyle = textCenterStyle
			rowNumStyle = numRightStyle
		}

		// Sheet 1 row data
		_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", orderRowIdx), i+1)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("B%d", orderRowIdx), inv)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("C%d", orderRowIdx), order.JobNumber)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("D%d", orderRowIdx), order.CreatedAt.Format("2006-01-02 15:04:05"))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("E%d", orderRowIdx), custName)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("F%d", orderRowIdx), custPhone)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("G%d", orderRowIdx), cashierName)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("H%d", orderRowIdx), mapOrderStatus(order.Status))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("I%d", orderRowIdx), mapPaymentStatus(order.PaymentStatus, order.GrandTotal, order.AmountPaid, hasRefund, isRefund))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("J%d", orderRowIdx), payMethod)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("K%d", orderRowIdx), order.GrandTotal)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("L%d", orderRowIdx), order.AmountPaid)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("M%d", orderRowIdx), sisa)

		_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", orderRowIdx), fmt.Sprintf("A%d", orderRowIdx), rowCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("B%d", orderRowIdx), fmt.Sprintf("D%d", orderRowIdx), rowCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("E%d", orderRowIdx), fmt.Sprintf("E%d", orderRowIdx), rowTextStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("F%d", orderRowIdx), fmt.Sprintf("J%d", orderRowIdx), rowCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("K%d", orderRowIdx), fmt.Sprintf("M%d", orderRowIdx), rowNumStyle)

		// Sheet 2 item rows
		for _, item := range order.OrderItems {
			catName := "-"
			prodName := "-"
			varName := "-"
			if item.ProductVariant != nil {
				varName = item.ProductVariant.VariantName
				prodName = item.ProductVariant.Product.Name
				if item.ProductVariant.Product.Category.Name != "" {
					catName = item.ProductVariant.Product.Category.Name
				}
			}

			ukuran := "-"
			if item.LengthM != nil && item.WidthM != nil && (*item.LengthM > 0 || *item.WidthM > 0) {
				ukuran = fmt.Sprintf("%.2f x %.2f m", *item.LengthM, *item.WidthM)
			}

			notes := "-"
			if item.ProductionNotes != nil && *item.ProductionNotes != "" {
				notes = *item.ProductionNotes
			}

			_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", itemRowIdx), itemSeq)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("B%d", itemRowIdx), inv)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("C%d", itemRowIdx), order.JobNumber)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("D%d", itemRowIdx), catName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("E%d", itemRowIdx), prodName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("F%d", itemRowIdx), varName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("G%d", itemRowIdx), ukuran)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("H%d", itemRowIdx), item.Quantity)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("I%d", itemRowIdx), item.UOM)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("J%d", itemRowIdx), item.PricePerUnit)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("K%d", itemRowIdx), notes)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("L%d", itemRowIdx), item.Subtotal)

			_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", itemRowIdx), fmt.Sprintf("C%d", itemRowIdx), rowCenterStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("D%d", itemRowIdx), fmt.Sprintf("F%d", itemRowIdx), rowTextStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("G%d", itemRowIdx), fmt.Sprintf("G%d", itemRowIdx), rowCenterStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("H%d", itemRowIdx), fmt.Sprintf("H%d", itemRowIdx), rowNumStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("I%d", itemRowIdx), fmt.Sprintf("I%d", itemRowIdx), rowCenterStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("J%d", itemRowIdx), fmt.Sprintf("J%d", itemRowIdx), rowNumStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("K%d", itemRowIdx), fmt.Sprintf("K%d", itemRowIdx), rowTextStyle)
			_ = f.SetCellStyle(sheet2, fmt.Sprintf("L%d", itemRowIdx), fmt.Sprintf("L%d", itemRowIdx), rowNumStyle)

			itemSeq++
			itemRowIdx++
		}
	}

	// Sheet 1 Total Row
	orderTotalRow := len(orders) + 5
	_ = f.MergeCell(sheet1, fmt.Sprintf("A%d", orderTotalRow), fmt.Sprintf("J%d", orderTotalRow))
	_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", orderTotalRow), "TOTAL")
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", orderTotalRow), fmt.Sprintf("J%d", orderTotalRow), totalLabelStyle)

	if len(orders) > 0 {
		_ = f.SetCellFormula(sheet1, fmt.Sprintf("K%d", orderTotalRow), fmt.Sprintf("SUM(K5:K%d)", orderTotalRow-1))
		_ = f.SetCellFormula(sheet1, fmt.Sprintf("L%d", orderTotalRow), fmt.Sprintf("SUM(L5:L%d)", orderTotalRow-1))
		_ = f.SetCellFormula(sheet1, fmt.Sprintf("M%d", orderTotalRow), fmt.Sprintf("SUM(M5:M%d)", orderTotalRow-1))
	} else {
		_ = f.SetCellValue(sheet1, fmt.Sprintf("K%d", orderTotalRow), 0)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("L%d", orderTotalRow), 0)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("M%d", orderTotalRow), 0)
	}
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("K%d", orderTotalRow), fmt.Sprintf("M%d", orderTotalRow), totalNumStyle)

	// Sheet 2 Total Row
	_ = f.MergeCell(sheet2, fmt.Sprintf("A%d", itemRowIdx), fmt.Sprintf("K%d", itemRowIdx))
	_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", itemRowIdx), "TOTAL")
	_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", itemRowIdx), fmt.Sprintf("K%d", itemRowIdx), totalLabelStyle)

	if itemSeq > 1 {
		_ = f.SetCellFormula(sheet2, fmt.Sprintf("L%d", itemRowIdx), fmt.Sprintf("SUM(L5:L%d)", itemRowIdx-1))
	} else {
		_ = f.SetCellValue(sheet2, fmt.Sprintf("L%d", itemRowIdx), 0)
	}
	_ = f.SetCellStyle(sheet2, fmt.Sprintf("L%d", itemRowIdx), fmt.Sprintf("L%d", itemRowIdx), totalNumStyle)

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
