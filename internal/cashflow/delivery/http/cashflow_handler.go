package http

import (
	"strconv"
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

	// Parsing Pagination
	pageStr := c.Query("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Parsing Optional Filters
	paymentMethod := c.Query("payment_method", "")
	flowType := c.Query("type", "")
	refType := c.Query("reference_type", "")
	cashierIDStr := c.Query("cashier_id", "")
	search := c.Query("search", "")

	var cashierID *uuid.UUID
	if cashierIDStr != "" {
		parsedID, err := uuid.Parse(cashierIDStr)
		if err == nil {
			cashierID = &parsedID
		}
	}

	filter := domain.CashFlowFilter{
		StartDate:     startDate,
		EndDate:       endDate,
		PaymentMethod: paymentMethod,
		Type:          flowType,
		ReferenceType: refType,
		CashierID:     cashierID,
		Search:        search,
		Page:          page,
		Limit:         limit,
	}

	report, total, err := h.usecase.GetReport(c.Context(), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch cash flow report", err.Error())
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  totalPages,
			CurrentPage: page,
			Limit:       limit,
		},
	}

	return response.Success(c, "Cash flow report fetched successfully", report, meta)
}

func (h *CashFlowHandler) GetSummary(c fiber.Ctx) error {
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

	// Parsing Optional Filters
	paymentMethod := c.Query("payment_method", "")
	flowType := c.Query("type", "")
	refType := c.Query("reference_type", "")
	cashierIDStr := c.Query("cashier_id", "")
	search := c.Query("search", "")

	var cashierID *uuid.UUID
	if cashierIDStr != "" {
		parsedID, err := uuid.Parse(cashierIDStr)
		if err == nil {
			cashierID = &parsedID
		}
	}

	filter := domain.CashFlowFilter{
		StartDate:     startDate,
		EndDate:       endDate,
		PaymentMethod: paymentMethod,
		Type:          flowType,
		ReferenceType: refType,
		CashierID:     cashierID,
		Search:        search,
	}

	summary, err := h.usecase.GetSummary(c.Context(), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch cash flow summary", err.Error())
	}

	return response.Success(c, "Cash flow summary fetched successfully", summary, nil)
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

func (h *CashFlowHandler) FindAllAccounts(c fiber.Ctx) error {
	accounts, err := h.usecase.FindAllAccounts(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch cash accounts", err.Error())
	}
	return response.Success(c, "Cash accounts fetched successfully", accounts, nil)
}
