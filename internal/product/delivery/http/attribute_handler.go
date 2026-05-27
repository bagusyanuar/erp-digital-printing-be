package http

import (
	"math"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AttributeHandler struct {
	attributeUsecase domain.AttributeUsecase
}

func NewAttributeHandler(attributeUsecase domain.AttributeUsecase) *AttributeHandler {
	return &AttributeHandler{attributeUsecase: attributeUsecase}
}

func (h *AttributeHandler) Create(c fiber.Ctx) error {
	var req dto.CreateAttributeReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	attribute := &domain.Attribute{
		Name:      req.Name,
		ValueType: req.ValueType,
	}

	if err := h.attributeUsecase.Create(c.Context(), attribute); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create attribute", err.Error())
	}

	return response.Created(c, "Attribute created successfully", attribute)
}

func (h *AttributeHandler) FindAll(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")

	attributes, total, err := h.attributeUsecase.FindAll(c.Context(), params, search)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch attributes", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Attributes fetched successfully", attributes, meta)
}

func (h *AttributeHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid attribute ID", err.Error())
	}

	attribute, err := h.attributeUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Attribute not found", err.Error())
	}

	return response.Success(c, "Attribute fetched successfully", attribute, nil)
}

func (h *AttributeHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid attribute ID", err.Error())
	}

	var req dto.UpdateAttributeReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	attribute, err := h.attributeUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Attribute not found", err.Error())
	}

	attribute.Name = req.Name
	attribute.ValueType = req.ValueType

	if err := h.attributeUsecase.Update(c.Context(), attribute); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update attribute", err.Error())
	}

	return response.Success(c, "Attribute updated successfully", attribute, nil)
}

func (h *AttributeHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid attribute ID", err.Error())
	}

	if err := h.attributeUsecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete attribute", err.Error())
	}

	return response.Success[any](c, "Attribute deleted successfully", nil, nil)
}
