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

	dateQuery := c.Query("date")
	startQuery := c.Query("start_date")
	endQuery := c.Query("end_date")

	if dateQuery != "" && startQuery == "" && endQuery == "" {
		if t, err := time.ParseInLocation("2006-01-02", dateQuery, time.Local); err == nil {
			s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			e := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
			start = &s
			end = &e
		}
	} else {
		if startQuery != "" {
			if t, err := time.ParseInLocation("2006-01-02", startQuery, time.Local); err == nil {
				s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
				start = &s
			}
		}
		if endQuery != "" {
			if t, err := time.ParseInLocation("2006-01-02", endQuery, time.Local); err == nil {
				e := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
				end = &e
			}
		}
	}

	f, err := h.reportUsecase.ExportSalesReport(c.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to export sales report", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "Failed to export report", err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			h.logger.Error("failed to close excel file", zap.Error(err))
		}
	}()

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to write excel to buffer", err.Error())
	}

	filename := fmt.Sprintf("Laporan_Penjualan_%s.xlsx", time.Now().Format("20060102150405"))
	if start != nil && end != nil {
		if start.Format("20060102") == end.Format("20060102") {
			filename = fmt.Sprintf("Laporan_Penjualan_%s.xlsx", start.Format("20060102"))
		} else {
			filename = fmt.Sprintf("Laporan_Penjualan_%s_%s.xlsx", start.Format("20060102"), end.Format("20060102"))
		}
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(buf.Bytes())
}

func (h *ReportHandler) ExportProfitLoss(c fiber.Ctx) error {
	var start, end *time.Time

	dateQuery := c.Query("date")
	startQuery := c.Query("start_date")
	endQuery := c.Query("end_date")

	if dateQuery != "" && startQuery == "" && endQuery == "" {
		if t, err := time.ParseInLocation("2006-01-02", dateQuery, time.Local); err == nil {
			s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			e := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
			start = &s
			end = &e
		}
	} else {
		if startQuery != "" {
			if t, err := time.ParseInLocation("2006-01-02", startQuery, time.Local); err == nil {
				s := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
				start = &s
			}
		}
		if endQuery != "" {
			if t, err := time.ParseInLocation("2006-01-02", endQuery, time.Local); err == nil {
				e := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
				end = &e
			}
		}
	}

	f, err := h.reportUsecase.ExportProfitLoss(c.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to export profit and loss report", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "Failed to export report", err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			h.logger.Error("failed to close excel file", zap.Error(err))
		}
	}()

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to write excel to buffer", err.Error())
	}

	filename := fmt.Sprintf("Laporan_Laba_Rugi_%s.xlsx", time.Now().Format("20060102150405"))
	if start != nil && end != nil {
		if start.Format("20060102") == end.Format("20060102") {
			filename = fmt.Sprintf("Laporan_Laba_Rugi_%s.xlsx", start.Format("20060102"))
		} else {
			filename = fmt.Sprintf("Laporan_Laba_Rugi_%s_%s.xlsx", start.Format("20060102"), end.Format("20060102"))
		}
	}

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
