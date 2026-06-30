package http

import (
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/dashboard/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	dashboardUsecase domain.DashboardUsecase
}

func NewDashboardHandler(dashboardUsecase domain.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{dashboardUsecase: dashboardUsecase}
}

func (h *DashboardHandler) GetWidgets(c fiber.Ctx) error {
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	var startDate, endDate *time.Time
	if (startDateStr == "" && endDateStr != "") || (startDateStr != "" && endDateStr == "") {
		return response.Error(c, fiber.StatusBadRequest, "Invalid date range", "Both start_date and end_date must be provided together")
	}

	if startDateStr != "" && endDateStr != "" {
		t1, err1 := time.Parse("2006-01-02", startDateStr)
		t2, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 != nil || err2 != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid date format", "Dates must be in YYYY-MM-DD format")
		}
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
		t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, t2.Location())
		startDate = &t1
		endDate = &t2
	}

	res, err := h.dashboardUsecase.GetWidgets(c.Context(), startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch dashboard widgets", err.Error())
	}

	return response.Success(c, "Dashboard widgets retrieved successfully", res, nil)
}
