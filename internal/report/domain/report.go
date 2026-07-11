package domain

import (
	"context"
	"time"

	"github.com/xuri/excelize/v2"
)

type ReportUsecase interface {
	ExportSalesReport(ctx context.Context, startDate *time.Time, endDate *time.Time) (*excelize.File, error)
}
