package domain

import (
	"context"
	"time"

	"github.com/xuri/excelize/v2"
)

type ProfitLossItem struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type ProfitLossSection struct {
	Title string           `json:"title"`
	Items []ProfitLossItem `json:"items"`
	Total float64          `json:"total"`
}

type ProfitLossGrossProfit struct {
	Amount           float64 `json:"amount"`
	MarginPercentage float64 `json:"margin_percentage"`
}

type ProfitLossNetProfit struct {
	Amount           float64 `json:"amount"`
	MarginPercentage float64 `json:"margin_percentage"`
	IsProfit         bool    `json:"is_profit"`
}

type ProfitLossRes struct {
	Period              string                `json:"period"`
	StartDate           string                `json:"start_date"`
	EndDate             string                `json:"end_date"`
	Revenue             ProfitLossSection     `json:"revenue"`
	COGS                ProfitLossSection     `json:"cogs"`
	GrossProfit         ProfitLossGrossProfit `json:"gross_profit"`
	OperationalExpenses ProfitLossSection     `json:"operational_expenses"`
	NetProfit           ProfitLossNetProfit   `json:"net_profit"`
}

type ProfitLossTrendItem struct {
	Month      string  `json:"month"`
	PeriodDate string  `json:"period_date"`
	Pendapatan float64 `json:"pendapatan"`
	HPP        float64 `json:"hpp"`
	Opex       float64 `json:"opex"`
	LabaBersih float64 `json:"laba_bersih"`
}

type ReportRepository interface {
	GetRevenueByCategory(ctx context.Context, startDate, endDate *time.Time) ([]ProfitLossItem, float64, error)
	GetCOGSByCategory(ctx context.Context, startDate, endDate *time.Time) ([]ProfitLossItem, float64, error)
	GetOperationalExpenses(ctx context.Context, startDate, endDate *time.Time) ([]ProfitLossItem, float64, error)
	GetMonthlyProfitLossTrend(ctx context.Context, months int) ([]ProfitLossTrendItem, error)
}

type ReportUsecase interface {
	ExportSalesReport(ctx context.Context, startDate *time.Time, endDate *time.Time) (*excelize.File, error)
	GetProfitLoss(ctx context.Context, startDate, endDate *time.Time) (*ProfitLossRes, error)
	GetProfitLossTrend(ctx context.Context, months int) ([]ProfitLossTrendItem, error)
}
