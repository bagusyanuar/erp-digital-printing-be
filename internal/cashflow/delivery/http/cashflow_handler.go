package http

import (
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CashFlowHandler struct {
	usecase domain.CashFlowUsecase
}

func NewCashFlowHandler(usecase domain.CashFlowUsecase) *CashFlowHandler {
	return &CashFlowHandler{usecase: usecase}
}

func (h *CashFlowHandler) GetReport(c fiber.Ctx) error {
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	if startDateStr == "" || endDateStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "Missing parameters", "Both start_date and end_date are required")
	}

	startDate, err1 := time.Parse("2006-01-02", startDateStr)
	endDate, err2 := time.Parse("2006-01-02", endDateStr)
	if err1 != nil || err2 != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid date format", "Dates must be in YYYY-MM-DD format")
	}

	// Set to start and end of day respectively
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	report, err := h.usecase.GetReport(c.Context(), startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch cash flow report", err.Error())
	}

	return response.Success(c, "Cash flow report fetched successfully", report, nil)
}

func (h *CashFlowHandler) CreateAdjustment(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	var req dto.CreateAdjustmentReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	cf, err := h.usecase.CreateAdjustment(c.Context(), cashierID, req.Amount, req.Type, req.PaymentMethod, req.Description)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create adjustment", err.Error())
	}

	res := dto.CashFlowRes{
		ID:              cf.ID,
		TransactionDate: cf.TransactionDate.Format("2006-01-02 15:04:05"),
		ReferenceType:   cf.ReferenceType,
		ReferenceID:     cf.ReferenceID,
		Type:            cf.Type,
		Amount:          cf.Amount,
		PaymentMethod:   cf.PaymentMethod,
		Description:     cf.Description,
		CashierID:       cf.CashierID,
	}

	return response.Created(c, "Cash flow adjustment created successfully", res)
}
