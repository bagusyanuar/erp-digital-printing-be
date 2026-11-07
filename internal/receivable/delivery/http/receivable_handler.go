package http

import (
	"math"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/receivable/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type ReceivableHandler struct {
	usecase domain.ReceivableUsecase
}

func NewReceivableHandler(usecase domain.ReceivableUsecase) *ReceivableHandler {
	return &ReceivableHandler{usecase: usecase}
}

func (h *ReceivableHandler) GetCustomerReceivables(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")
	status := c.Query("status", "ALL")
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	var startDate, endDate *time.Time
	if startDateStr != "" && endDateStr != "" {
		t1, err1 := time.ParseInLocation("2006-01-02", startDateStr, time.Local)
		t2, err2 := time.ParseInLocation("2006-01-02", endDateStr, time.Local)
		if err1 == nil && err2 == nil {
			t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local)
			t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, time.Local)
			startDate = &t1
			endDate = &t2
		}
	}

	data, total, err := h.usecase.GetCustomerReceivables(c.Context(), params, search, status, startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch customer receivables", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Customer receivables retrieved successfully", data, meta)
}

func (h *ReceivableHandler) GetSettledInvoices(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	var startDate, endDate *time.Time
	if startDateStr != "" && endDateStr != "" {
		t1, err1 := time.ParseInLocation("2006-01-02", startDateStr, time.Local)
		t2, err2 := time.ParseInLocation("2006-01-02", endDateStr, time.Local)
		if err1 == nil && err2 == nil {
			t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local)
			t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, time.Local)
			startDate = &t1
			endDate = &t2
		}
	}

	data, total, err := h.usecase.GetSettledInvoices(c.Context(), params, search, startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch settled receivables", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Settled receivables retrieved successfully", data, meta)
}

func (h *ReceivableHandler) GetSummary(c fiber.Ctx) error {
	res, err := h.usecase.GetSummary(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch receivable summary", err.Error())
	}

	return response.Success(c, "Receivable summary retrieved successfully", res, nil)
}
