package http

import (
	"math"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ResellerHandler struct {
	resellerUsecase domain.ResellerUsecase
}

func NewResellerHandler(resellerUsecase domain.ResellerUsecase) *ResellerHandler {
	return &ResellerHandler{resellerUsecase: resellerUsecase}
}

func (h *ResellerHandler) Create(c fiber.Ctx) error {
	var req dto.CreateResellerReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	reseller := &domain.Reseller{
		CustomerLevelID: req.CustomerLevelID,
		Name:            req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
		CreditLimit:     req.CreditLimit,
	}

	if err := h.resellerUsecase.Create(c.Context(), reseller); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create reseller", err.Error())
	}

	return response.Created(c, "Reseller created successfully", reseller)
}

func (h *ResellerHandler) FindAll(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")

	resellers, total, err := h.resellerUsecase.FindAll(c.Context(), params, search)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch resellers", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Resellers fetched successfully", resellers, meta)
}

func (h *ResellerHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid reseller ID", err.Error())
	}

	reseller, err := h.resellerUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Reseller not found", err.Error())
	}

	return response.Success(c, "Reseller fetched successfully", reseller, nil)
}

func (h *ResellerHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid reseller ID", err.Error())
	}

	var req dto.UpdateResellerReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	reseller, err := h.resellerUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Reseller not found", err.Error())
	}

	reseller.CustomerLevelID = req.CustomerLevelID
	reseller.Name = req.Name
	reseller.Email = req.Email
	reseller.Phone = req.Phone
	reseller.Address = req.Address
	reseller.CreditLimit = req.CreditLimit

	if err := h.resellerUsecase.Update(c.Context(), reseller); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update reseller", err.Error())
	}

	return response.Success(c, "Reseller updated successfully", reseller, nil)
}

func (h *ResellerHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid reseller ID", err.Error())
	}

	if err := h.resellerUsecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete reseller", err.Error())
	}

	return response.Success[any](c, "Reseller deleted successfully", nil, nil)
}
