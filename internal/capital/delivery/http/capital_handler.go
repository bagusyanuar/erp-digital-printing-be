package http

import (
	"strconv"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/capital/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/capital/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CapitalHandler struct {
	usecase domain.CapitalUsecase
}

func NewCapitalHandler(usecase domain.CapitalUsecase) *CapitalHandler {
	return &CapitalHandler{usecase: usecase}
}

func (h *CapitalHandler) Create(c fiber.Ctx) error {
	creatorID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found in context", nil)
	}

	var req dto.CreateCapitalReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	tx, err := h.usecase.Create(c.Context(), creatorID, req.Type, req.Amount, req.PaymentMethod, req.Description)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create capital transaction", err.Error())
	}

	res := dto.CapitalTransactionRes{
		ID:              tx.ID,
		TransactionDate: tx.TransactionDate.Format("2006-01-02 15:04:05"),
		Type:            tx.Type,
		Amount:          tx.Amount,
		PaymentMethod:   tx.PaymentMethod,
		Description:     tx.Description,
		CreatedBy:       tx.CreatedBy,
		CreatedAt:       tx.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return response.Created(c, "Capital transaction created successfully", res)
}

func (h *CapitalHandler) FindAll(c fiber.Ctx) error {
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

	txType := c.Query("type", "")
	search := c.Query("search", "")

	var startDate *time.Time
	var endDate *time.Time

	startDateStr := c.Query("start_date", "")
	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			sDate := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
			startDate = &sDate
		}
	}

	endDateStr := c.Query("end_date", "")
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			eDate := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
			endDate = &eDate
		}
	}

	filter := domain.CapitalFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Type:      txType,
		Search:    search,
		Page:      page,
		Limit:     limit,
	}

	transactions, total, err := h.usecase.FindAll(c.Context(), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch capital transactions", err.Error())
	}

	resList := make([]dto.CapitalTransactionRes, 0, len(transactions))
	for _, tx := range transactions {
		creatorName := "System"
		if tx.Creator != nil {
			creatorName = tx.Creator.Username
		}

		resList = append(resList, dto.CapitalTransactionRes{
			ID:              tx.ID,
			TransactionDate: tx.TransactionDate.Format("2006-01-02 15:04:05"),
			Type:            tx.Type,
			Amount:          tx.Amount,
			PaymentMethod:   tx.PaymentMethod,
			Description:     tx.Description,
			CreatedBy:       tx.CreatedBy,
			CreatorName:     creatorName,
			CreatedAt:       tx.CreatedAt.Format("2006-01-02 15:04:05"),
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

	return response.Success(c, "Capital transactions fetched successfully", resList, meta)
}

func (h *CapitalHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := h.usecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to cancel capital transaction", err.Error())
	}

	return response.Success[any](c, "Capital transaction cancelled successfully", nil, nil)
}
