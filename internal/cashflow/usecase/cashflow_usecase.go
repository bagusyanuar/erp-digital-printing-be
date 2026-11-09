package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type cashFlowUsecase struct {
	repo domain.CashFlowRepository
	db   *gorm.DB
}

func NewCashFlowUsecase(repo domain.CashFlowRepository, db *gorm.DB) domain.CashFlowUsecase {
	return &cashFlowUsecase{repo: repo, db: db}
}

func mapCashFlowCategory(refType string) string {
	switch strings.ToUpper(refType) {
	case domain.RefOrderPayment:
		return "Penjualan"
	case "CAPITAL_INJECTION", domain.RefCapital:
		return "Setoran Modal"
	case "EXPENSE", "EXPENSE_PAYMENT":
		return "Pengeluaran"
	case domain.RefRefund:
		return "Refund"
	case domain.RefAdjustment:
		return "Penyesuaian Kas"
	case "CAPITAL_WITHDRAWAL":
		return "Penarikan Owner (Prive)"
	case domain.RefFundTransfer:
		return "Pemindahan Dana"
	default:
		if refType == "" {
			return "-"
		}
		return refType
	}
}

func (u *cashFlowUsecase) GetReport(ctx context.Context, filter domain.CashFlowFilter) ([]domain.CashFlowTransactionRes, int64, error) {
	cashFlows, total, err := u.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]domain.CashFlowTransactionRes, 0, len(cashFlows))

	for _, cf := range cashFlows {
		cashierName := "System"
		if cf.Cashier != nil {
			cashierName = cf.Cashier.Username
		}

		transactions = append(transactions, domain.CashFlowTransactionRes{
			ID:              cf.ID,
			TransactionDate: cf.TransactionDate,
			ReferenceType:   cf.ReferenceType,
			ReferenceID:     cf.ReferenceID,
			Type:            cf.Type,
			Amount:          cf.Amount,
			PaymentMethod:   cf.PaymentMethod,
			Description:     cf.Description,
			CustomerName:    cf.CustomerName,
			InvoiceNumber:   cf.InvoiceNumber,
			CashierName:     cashierName,
		})
	}

	return transactions, total, nil
}

func (u *cashFlowUsecase) CreateAdjustment(ctx context.Context, cashierID uuid.UUID, amount float64, flowType string, paymentMethod string, description string) (*domain.CashFlow, error) {
	desc := description
	cf := &domain.CashFlow{
		ID:              uuid.New(),
		TransactionDate: time.Now(),
		ReferenceType:   domain.RefAdjustment,
		ReferenceID:     nil,
		Type:            flowType,
		Amount:          amount,
		PaymentMethod:   paymentMethod,
		Description:     &desc,
		CashierID:       cashierID,
	}

	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := u.repo.FindAccountByNameWithLock(ctx, tx, paymentMethod)
		if err != nil {
			return err
		}

		if flowType == domain.TypeDebit {
			acc.Balance += amount
		} else {
			acc.Balance -= amount
		}

		if err := u.repo.UpdateAccount(ctx, tx, acc); err != nil {
			return err
		}

		if err := u.repo.CreateTx(ctx, tx, cf); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (u *cashFlowUsecase) FindAllAccounts(ctx context.Context) ([]domain.CashAccount, error) {
	return u.repo.FindAllAccounts(ctx)
}

func (u *cashFlowUsecase) GetSummary(ctx context.Context, filter domain.CashFlowFilter) (*domain.CashFlowSummaryRes, error) {
	return u.repo.GetSummary(ctx, filter)
}

func (u *cashFlowUsecase) ExportExcel(ctx context.Context, filter domain.CashFlowFilter) (*excelize.File, error) {
	// 1. Fetch transactions (no pagination limit for export)
	filter.Page = 1
	filter.Limit = 100000
	cashFlows, _, err := u.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cash flows: %w", err)
	}

	// 2. Fetch summary
	summaryRes, err := u.repo.GetSummary(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cash flow summary: %w", err)
	}

	// 3. Chronological order (ASC: oldest first so running balance accumulates naturally)
	for i, j := 0, len(cashFlows)-1; i < j; i, j = i+1, j-1 {
		cashFlows[i], cashFlows[j] = cashFlows[j], cashFlows[i]
	}

	f := excelize.NewFile()

	sheet1 := "Mutasi Kas"
	sheet2 := "Ringkasan Saldo"

	_ = f.SetSheetName("Sheet1", sheet1)
	_, err = f.NewSheet(sheet2)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(0)

	// Period text
	var periodeText string
	if filter.StartDate.Format("2006-01-02") == filter.EndDate.Format("2006-01-02") {
		periodeText = fmt.Sprintf("Periode: %s", filter.StartDate.Format("02/01/2006"))
	} else {
		periodeText = fmt.Sprintf("Periode: %s s/d %s", filter.StartDate.Format("02/01/2006"), filter.EndDate.Format("02/01/2006"))
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

	debitNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Color: "059669"},
		CustomNumFmt: &numFmt,
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
	})

	creditNumStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Color: "DC2626"},
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

	// -------------------------------------------------------------
	// Sheet 1: Mutasi Kas
	// -------------------------------------------------------------
	_ = f.SetRowHeight(sheet1, 1, 24)
	_ = f.SetRowHeight(sheet1, 2, 18)
	_ = f.SetRowHeight(sheet1, 4, 26)

	_ = f.SetCellValue(sheet1, "A1", "LAPORAN ARUS KAS")
	_ = f.SetCellStyle(sheet1, "A1", "A1", titleStyle)
	_ = f.SetCellValue(sheet1, "A2", periodeText)
	_ = f.SetCellStyle(sheet1, "A2", "A2", periodStyle)

	sheet1Headers := []string{
		"No.", "Waktu / Tanggal", "Kategori", "Deskripsi", "No. Invoice / Ref",
		"Pelanggan", "Metode / Akun", "Kasir", "Kas Masuk (Debit)", "Kas Keluar (Kredit)", "Saldo Berjalan",
	}

	for i, h := range sheet1Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet1, cell, h)
		_ = f.SetCellStyle(sheet1, cell, cell, headerStyle)
	}

	_ = f.SetColWidth(sheet1, "A", "A", 6)
	_ = f.SetColWidth(sheet1, "B", "B", 18)
	_ = f.SetColWidth(sheet1, "C", "C", 22)
	_ = f.SetColWidth(sheet1, "D", "D", 32)
	_ = f.SetColWidth(sheet1, "E", "E", 20)
	_ = f.SetColWidth(sheet1, "F", "F", 22)
	_ = f.SetColWidth(sheet1, "G", "H", 16)
	_ = f.SetColWidth(sheet1, "I", "K", 20)

	runningBalance := 0.0
	for i, cf := range cashFlows {
		row := i + 5

		desc := "-"
		if cf.Description != nil && *cf.Description != "" {
			desc = *cf.Description
		}

		invRef := "-"
		if cf.InvoiceNumber != nil && *cf.InvoiceNumber != "" {
			invRef = *cf.InvoiceNumber
		} else if cf.ReferenceID != nil {
			invRef = cf.ReferenceID.String()[:8]
		}

		custName := "-"
		if cf.CustomerName != nil && *cf.CustomerName != "" {
			custName = *cf.CustomerName
		}

		cashierName := "System"
		if cf.Cashier != nil && cf.Cashier.Username != "" {
			cashierName = cf.Cashier.Username
		}

		debit := 0.0
		credit := 0.0
		if strings.ToUpper(cf.Type) == domain.TypeDebit {
			debit = cf.Amount
			runningBalance += cf.Amount
		} else {
			credit = cf.Amount
			runningBalance -= cf.Amount
		}

		_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), i+1)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), cf.TransactionDate.Format("2006-01-02 15:04:05"))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("C%d", row), mapCashFlowCategory(cf.ReferenceType))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("D%d", row), desc)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("E%d", row), invRef)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("F%d", row), custName)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("G%d", row), strings.ToUpper(cf.PaymentMethod))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("H%d", row), cashierName)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("I%d", row), debit)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("J%d", row), credit)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("K%d", row), runningBalance)

		_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), textLeftStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), textLeftStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), textLeftStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("G%d", row), fmt.Sprintf("H%d", row), textCenterStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), debitNumStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), creditNumStyle)
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), numRightStyle)
	}

	// Sheet 1 Total Row
	sheet1TotalRow := len(cashFlows) + 5
	_ = f.MergeCell(sheet1, fmt.Sprintf("A%d", sheet1TotalRow), fmt.Sprintf("H%d", sheet1TotalRow))
	_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", sheet1TotalRow), "TOTAL")
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", sheet1TotalRow), fmt.Sprintf("H%d", sheet1TotalRow), totalLabelStyle)

	if len(cashFlows) > 0 {
		_ = f.SetCellFormula(sheet1, fmt.Sprintf("I%d", sheet1TotalRow), fmt.Sprintf("SUM(I5:I%d)", sheet1TotalRow-1))
		_ = f.SetCellFormula(sheet1, fmt.Sprintf("J%d", sheet1TotalRow), fmt.Sprintf("SUM(J5:J%d)", sheet1TotalRow-1))
		_ = f.SetCellFormula(sheet1, fmt.Sprintf("K%d", sheet1TotalRow), fmt.Sprintf("I%d-J%d", sheet1TotalRow, sheet1TotalRow))
	} else {
		_ = f.SetCellValue(sheet1, fmt.Sprintf("I%d", sheet1TotalRow), 0)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("J%d", sheet1TotalRow), 0)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("K%d", sheet1TotalRow), 0)
	}
	_ = f.SetCellStyle(sheet1, fmt.Sprintf("I%d", sheet1TotalRow), fmt.Sprintf("K%d", sheet1TotalRow), totalNumStyle)

	// -------------------------------------------------------------
	// Sheet 2: Ringkasan Saldo (Metode Pembayaran: Cash, QRIS, Transfer)
	// -------------------------------------------------------------
	_ = f.SetRowHeight(sheet2, 1, 24)
	_ = f.SetRowHeight(sheet2, 2, 18)
	_ = f.SetRowHeight(sheet2, 4, 26)

	_ = f.SetCellValue(sheet2, "A1", "RINGKASAN ARUS KAS PER METODE PEMBAYARAN")
	_ = f.SetCellStyle(sheet2, "A1", "A1", titleStyle)
	_ = f.SetCellValue(sheet2, "A2", periodeText)
	_ = f.SetCellStyle(sheet2, "A2", "A2", periodStyle)

	sheet2Headers := []string{
		"No.", "Metode Pembayaran / Akun", "Total Kas Masuk (Debit)", "Total Kas Keluar (Kredit)", "Saldo Bersih (Net)",
	}

	for i, h := range sheet2Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet2, cell, h)
		_ = f.SetCellStyle(sheet2, cell, cell, headerStyle)
	}

	_ = f.SetColWidth(sheet2, "A", "A", 6)
	_ = f.SetColWidth(sheet2, "B", "B", 28)
	_ = f.SetColWidth(sheet2, "C", "E", 25)

	type MethodData struct {
		Name    string
		Debit   float64
		Credit  float64
		Balance float64
	}

	// Prepare standard methods list: CASH, QRIS, TRANSFER + any others from summary
	methodMap := make(map[string]domain.CashFlowMethodDetail)
	if summaryRes != nil && summaryRes.DetailsByMethod != nil {
		for k, v := range summaryRes.DetailsByMethod {
			methodMap[strings.ToUpper(k)] = v
		}
	}

	standardOrder := []string{"CASH", "QRIS", "TRANSFER"}
	seenMethods := make(map[string]bool)
	var finalMethods []MethodData

	for _, m := range standardOrder {
		seenMethods[m] = true
		detail := methodMap[m]
		displayName := m
		if m == "CASH" {
			displayName = "KAS TUNAI (CASH)"
		} else if m == "TRANSFER" {
			displayName = "TRANSFER BANK"
		}
		finalMethods = append(finalMethods, MethodData{
			Name:    displayName,
			Debit:   detail.Debit,
			Credit:  detail.Credit,
			Balance: detail.Balance,
		})
	}

	// Add other accounts if present in DB
	for k, detail := range methodMap {
		if !seenMethods[k] {
			finalMethods = append(finalMethods, MethodData{
				Name:    k,
				Debit:   detail.Debit,
				Credit:  detail.Credit,
				Balance: detail.Balance,
			})
		}
	}

	for i, m := range finalMethods {
		row := i + 5
		_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), i+1)
		_ = f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), m.Name)
		_ = f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), m.Debit)
		_ = f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), m.Credit)
		_ = f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), m.Balance)

		_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), textCenterStyle)
		_ = f.SetCellStyle(sheet2, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), textLeftStyle)
		_ = f.SetCellStyle(sheet2, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), debitNumStyle)
		_ = f.SetCellStyle(sheet2, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), creditNumStyle)
		_ = f.SetCellStyle(sheet2, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), numRightStyle)
	}

	// Sheet 2 Total Row
	sheet2TotalRow := len(finalMethods) + 5
	_ = f.MergeCell(sheet2, fmt.Sprintf("A%d", sheet2TotalRow), fmt.Sprintf("B%d", sheet2TotalRow))
	_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", sheet2TotalRow), "TOTAL KESELURUHAN")
	_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", sheet2TotalRow), fmt.Sprintf("B%d", sheet2TotalRow), totalLabelStyle)

	_ = f.SetCellFormula(sheet2, fmt.Sprintf("C%d", sheet2TotalRow), fmt.Sprintf("SUM(C5:C%d)", sheet2TotalRow-1))
	_ = f.SetCellFormula(sheet2, fmt.Sprintf("D%d", sheet2TotalRow), fmt.Sprintf("SUM(D5:D%d)", sheet2TotalRow-1))
	_ = f.SetCellFormula(sheet2, fmt.Sprintf("E%d", sheet2TotalRow), fmt.Sprintf("C%d-D%d", sheet2TotalRow, sheet2TotalRow))
	_ = f.SetCellStyle(sheet2, fmt.Sprintf("C%d", sheet2TotalRow), fmt.Sprintf("E%d", sheet2TotalRow), totalNumStyle)

	return f, nil
}

