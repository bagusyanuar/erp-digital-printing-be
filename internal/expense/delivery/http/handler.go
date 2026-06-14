package http

import (
	"math"
	"strconv"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ExpenseHandler struct {
	usecase domain.ExpenseUsecase
}

func NewExpenseHandler(usecase domain.ExpenseUsecase) *ExpenseHandler {
	return &ExpenseHandler{usecase: usecase}
}

// --- CATEGORIES ---

func (h *ExpenseHandler) CreateCategory(c fiber.Ctx) error {
	var req dto.CreateCategoryReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	category := &domain.ExpenseCategory{
		Name:              req.Name,
		Group:             req.Group,
		ProductCategoryID: req.ProductCategoryID,
	}

	if err := h.usecase.CreateCategory(c.Context(), category); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create expense category", err.Error())
	}

	return response.Created(c, "Expense category created successfully", category)
}

func (h *ExpenseHandler) FindAllCategories(c fiber.Ctx) error {
	group := c.Query("group", "")
	categories, err := h.usecase.FindAllCategories(c.Context(), group)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch expense categories", err.Error())
	}

	return response.Success(c, "Expense categories fetched successfully", categories, nil)
}

func (h *ExpenseHandler) FindCategoryByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid category ID", err.Error())
	}

	category, err := h.usecase.FindCategoryByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Expense category not found", err.Error())
	}

	return response.Success(c, "Expense category fetched successfully", category, nil)
}

func (h *ExpenseHandler) UpdateCategory(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid category ID", err.Error())
	}

	var req dto.UpdateCategoryReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	category, err := h.usecase.FindCategoryByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Expense category not found", err.Error())
	}

	category.Name = req.Name
	category.Group = req.Group
	category.ProductCategoryID = req.ProductCategoryID

	if err := h.usecase.UpdateCategory(c.Context(), category); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update expense category", err.Error())
	}

	return response.Success(c, "Expense category updated successfully", category, nil)
}

func (h *ExpenseHandler) DeleteCategory(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid category ID", err.Error())
	}

	if err := h.usecase.DeleteCategory(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete expense category", err.Error())
	}

	return response.Success[any](c, "Expense category deleted successfully", nil, nil)
}

// --- EXPENSES ---

func (h *ExpenseHandler) CreateExpense(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	var req dto.CreateExpenseReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var expenseDate time.Time
	if req.ExpenseDate != nil {
		expenseDate = *req.ExpenseDate
	} else {
		expenseDate = time.Now()
	}

	expense := &domain.Expense{
		ExpenseCategoryID: req.ExpenseCategoryID,
		Amount:            req.Amount,
		ExpenseDate:       expenseDate,
		PaymentMethod:     req.PaymentMethod,
		Description:       req.Description,
		CashierID:         cashierID,
	}

	if err := h.usecase.CreateExpense(c.Context(), expense); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create expense", err.Error())
	}

	return response.Created(c, "Expense recorded successfully", expense)
}

func (h *ExpenseHandler) FindAllExpenses(c fiber.Ctx) error {
	var startDate, endDate *time.Time
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	if startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid start_date format", err.Error())
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		startDate = &t
	}

	if endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid end_date format", err.Error())
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		endDate = &t
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

	group := c.Query("group", "")
	categoryIDStr := c.Query("expense_category_id", "")
	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		parsed, err := uuid.Parse(categoryIDStr)
		if err == nil {
			categoryID = &parsed
		}
	}

	search := c.Query("search", "")

	filter := domain.ExpenseFilter{
		StartDate:  startDate,
		EndDate:    endDate,
		Group:      group,
		CategoryID: categoryID,
		Search:     search,
		Page:       page,
		Limit:      limit,
	}

	expenses, total, err := h.usecase.FindAllExpenses(c.Context(), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch expenses", err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  totalPages,
			CurrentPage: page,
			Limit:       limit,
		},
	}

	return response.Success(c, "Expenses fetched successfully", expenses, meta)
}

func (h *ExpenseHandler) DeleteExpense(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid expense ID", err.Error())
	}

	if err := h.usecase.DeleteExpense(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete expense", err.Error())
	}

	return response.Success[any](c, "Expense deleted successfully", nil, nil)
}

// --- ANALYTICS ---

func (h *ExpenseHandler) GetSummary(c fiber.Ctx) error {
	var startDate, endDate *time.Time
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	if startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid start_date format", err.Error())
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		startDate = &t
	}

	if endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid end_date format", err.Error())
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		endDate = &t
	}

	summary, err := h.usecase.GetSummary(c.Context(), startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch expense summary", err.Error())
	}

	return response.Success(c, "Expense summary fetched successfully", summary, nil)
}

func (h *ExpenseHandler) GetByProductCategory(c fiber.Ctx) error {
	var startDate, endDate *time.Time
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	if startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid start_date format", err.Error())
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		startDate = &t
	}

	if endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid end_date format", err.Error())
		}
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		endDate = &t
	}

	byProductCategory, err := h.usecase.GetByProductCategory(c.Context(), startDate, endDate)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch expense by product category", err.Error())
	}

	return response.Success(c, "Expense by product category fetched successfully", byProductCategory, nil)
}
