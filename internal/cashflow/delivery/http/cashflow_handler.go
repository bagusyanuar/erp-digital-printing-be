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
	usecase             domain.CashFlowUsecase
	fundTransferUsecase domain.FundTransferUsecase
}

func NewCashFlowHandler(usecase domain.CashFlowUsecase, fundTransferUsecase domain.FundTransferUsecase) *CashFlowHandler {
	return &CashFlowHandler{usecase: usecase, fundTransferUsecase: fundTransferUsecase}
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
		InvoiceNumber:   cf.InvoiceNumber,
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

func (h *CashFlowHandler) CreateFundTransfer(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	var req dto.CreateFundTransferReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var transferDate *time.Time
	if req.TransferDate != "" {
		var parsedDate time.Time
		var err error
		if len(req.TransferDate) > 10 {
			parsedDate, err = time.ParseInLocation("2006-01-02 15:04:05", req.TransferDate, time.Local)
		} else {
			parsedDate, err = time.ParseInLocation("2006-01-02", req.TransferDate, time.Local)
		}
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid transfer_date format", "Date must be YYYY-MM-DD or YYYY-MM-DD HH:mm:ss")
		}
		transferDate = &parsedDate
	}

	transfer, err := h.fundTransferUsecase.Transfer(c.Context(), cashierID, req.FromAccount, req.ToAccount, req.Amount, req.Notes, transferDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to execute fund transfer", err.Error())
	}

	res := dto.FundTransferRes{
		ID:           transfer.ID,
		TransferDate: transfer.TransferDate.Format("2006-01-02 15:04:05"),
		FromAccount:  req.FromAccount,
		ToAccount:    req.ToAccount,
		Amount:       transfer.Amount,
		Notes:        transfer.Notes,
		CreatedAt:    transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return response.Created(c, "Fund transfer executed successfully", res)
}

func (h *CashFlowHandler) GetFundTransfers(c fiber.Ctx) error {
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	var startDate, endDate time.Time
	if startDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
		}
	}
	if endDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, parsedDate.Location())
		}
	}

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

	filter := domain.FundTransferFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
	}

	transfers, total, err := h.fundTransferUsecase.FindAll(c.Context(), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch fund transfers", err.Error())
	}

	resList := make([]dto.FundTransferRes, 0, len(transfers))
	for _, t := range transfers {
		cashierName := "System"
		if t.Cashier != nil {
			cashierName = t.Cashier.Username
		}
		fromName := ""
		if t.FromAccount != nil {
			fromName = t.FromAccount.Name
		}
		toName := ""
		if t.ToAccount != nil {
			toName = t.ToAccount.Name
		}

		resList = append(resList, dto.FundTransferRes{
			ID:           t.ID,
			TransferDate: t.TransferDate.Format("2006-01-02 15:04:05"),
			FromAccount:  fromName,
			ToAccount:    toName,
			Amount:       t.Amount,
			Notes:        t.Notes,
			CashierName:  cashierName,
			CreatedAt:    t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
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

	return response.Success(c, "Fund transfers fetched successfully", resList, meta)
}

func (h *CashFlowHandler) CancelFundTransfer(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	err = h.fundTransferUsecase.Cancel(c.Context(), cashierID, id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to cancel fund transfer", err.Error())
	}

	return response.Success[any](c, "Fund transfer cancelled successfully", nil, nil)
}

func (h *CashFlowHandler) GetFundTransferWidgets(c fiber.Ctx) error {
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	var startDate, endDate time.Time
	if startDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
		}
	}
	if endDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, parsedDate.Location())
		}
	}

	widgets, err := h.fundTransferUsecase.GetWidgets(c.Context(), startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch fund transfer widgets", err.Error())
	}

	return response.Success(c, "Fund transfer widgets fetched successfully", widgets, nil)
}
