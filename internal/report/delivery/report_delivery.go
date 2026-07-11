package delivery

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/report/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type ReportHandler struct {
	reportUsecase domain.ReportUsecase
	logger        *zap.Logger
}

func NewReportHandler(reportUsecase domain.ReportUsecase, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		reportUsecase: reportUsecase,
		logger:        logger,
	}
}

func (h *ReportHandler) ExportSalesReport(c fiber.Ctx) error {
	var start, end *time.Time
	
	startQuery := c.Query("start_date")
	endQuery := c.Query("end_date")

	if startQuery != "" {
		if t, err := time.Parse("2006-01-02", startQuery); err == nil {
			start = &t
		}
	}
	if endQuery != "" {
		if t, err := time.Parse("2006-01-02", endQuery); err == nil {
			// Set ke akhir hari (23:59:59)
			endOfDay := t.Add(time.Hour*23 + time.Minute*59 + time.Second*59)
			end = &endOfDay
		}
	}

	f, err := h.reportUsecase.ExportSalesReport(c.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to export sales report", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "Failed to export report", err.Error())
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to write excel to buffer", err.Error())
	}

	filename := fmt.Sprintf("Report_Sales_%s.xlsx", time.Now().Format("20060102150405"))

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	
	return c.Send(buf.Bytes())
}
