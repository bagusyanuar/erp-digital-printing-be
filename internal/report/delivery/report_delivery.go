package delivery

import (
	"bytes"
	"fmt"
	"strconv"
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

func (h *ReportHandler) GetProfitLoss(c fiber.Ctx) error {
	var start, end *time.Time
	startQuery := c.Query("start_date")
	endQuery := c.Query("end_date")

	if startQuery != "" {
		if t, err := time.ParseInLocation("2006-01-02", startQuery, time.Local); err == nil {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			start = &t
		} else {
			return response.Error(c, fiber.StatusBadRequest, "Invalid start_date format", "Date must be in YYYY-MM-DD format")
		}
	}
	if endQuery != "" {
		if t, err := time.ParseInLocation("2006-01-02", endQuery, time.Local); err == nil {
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
			end = &t
		} else {
			return response.Error(c, fiber.StatusBadRequest, "Invalid end_date format", "Date must be in YYYY-MM-DD format")
		}
	}

	res, err := h.reportUsecase.GetProfitLoss(c.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to fetch profit and loss report", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch profit and loss report", err.Error())
	}

	return response.Success(c, "Profit loss report fetched successfully", res, nil)
}

func (h *ReportHandler) GetProfitLossTrend(c fiber.Ctx) error {
	monthsQuery := c.Query("months", "6")
	months, err := strconv.Atoi(monthsQuery)
	if err != nil || months <= 0 {
		months = 6
	}

	res, err := h.reportUsecase.GetProfitLossTrend(c.Context(), months)
	if err != nil {
		h.logger.Error("Failed to fetch profit and loss trend", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch profit and loss trend", err.Error())
	}

	return response.Success(c, "Profit loss trend fetched successfully", res, nil)
}
