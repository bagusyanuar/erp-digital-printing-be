package http

import (
	"math"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/payable/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type PayableHandler struct {
	usecase domain.PayableUsecase
}

func NewPayableHandler(usecase domain.PayableUsecase) *PayableHandler {
	return &PayableHandler{usecase: usecase}
}

type PayExpenseReq struct {
	Payments []PayExpenseItemReq `json:"payments"`
	Notes    *string             `json:"notes,omitempty"`
}

type PayExpenseItemReq struct {
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	SourceAccount *string `json:"source_account,omitempty"`
}

func (h *PayableHandler) GetVendorPayables(c fiber.Ctx) error {
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

	data, total, err := h.usecase.GetVendorPayables(c.Context(), params, search, status, startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch vendor payables", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Vendor payables retrieved successfully", data, meta)
}

func (h *PayableHandler) GetSettledPayables(c fiber.Ctx) error {
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

	data, total, err := h.usecase.GetSettledPayables(c.Context(), params, search, startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch settled payables", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Settled payables retrieved successfully", data, meta)
}

func (h *PayableHandler) GetSummary(c fiber.Ctx) error {
	res, err := h.usecase.GetSummary(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch payable summary", err.Error())
	}

	return response.Success(c, "Payable summary retrieved successfully", res, nil)
}

func (h *PayableHandler) PayExpense(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	idStr := c.Params("id")
	expenseID, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid expense ID", err.Error())
	}

	var req PayExpenseReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	items := make([]domain.PayExpenseItemInput, len(req.Payments))
	for i, p := range req.Payments {
		items[i] = domain.PayExpenseItemInput{
			PaymentMethod: p.PaymentMethod,
			Amount:        p.Amount,
			SourceAccount: p.SourceAccount,
		}
	}

	input := domain.PayVendorExpenseInput{
		ExpenseID: expenseID,
		CashierID: cashierID,
		Payments:  items,
		Notes:     req.Notes,
	}

	vendor, err := h.usecase.PayExpense(c.Context(), input)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to record expense payment", err.Error())
	}

	return response.Success(c, "Expense payment recorded successfully", vendor, nil)
}

