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

	var options []domain.AttributeOption
	if req.ValueType == "options" {
		for _, optVal := range req.Options {
			if optVal != "" {
				options = append(options, domain.AttributeOption{
					Value: optVal,
				})
			}
		}
	}

	attribute := &domain.Attribute{
		Name:      req.Name,
		ValueType: req.ValueType,
		Options:   options,
	}

	if err := h.attributeUsecase.Create(c.Context(), attribute); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create attribute", err.Error())
	}

	res := mapToAttributeRes(attribute)
	return response.Created(c, "Attribute created successfully", res)
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

	resList := make([]dto.AttributeRes, 0)
	for i := range attributes {
		resList = append(resList, mapToAttributeRes(&attributes[i]))
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Attributes fetched successfully", resList, meta)
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

	res := mapToAttributeRes(attribute)
	return response.Success(c, "Attribute fetched successfully", res, nil)
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

	var options []domain.AttributeOption
	if req.ValueType == "options" {
		for _, optVal := range req.Options {
			if optVal != "" {
				options = append(options, domain.AttributeOption{
					AttributeID: attribute.ID,
					Value:       optVal,
				})
			}
		}
	}

	attribute.Name = req.Name
	attribute.ValueType = req.ValueType
	attribute.Options = options

	if err := h.attributeUsecase.Update(c.Context(), attribute); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update attribute", err.Error())
	}

	res := mapToAttributeRes(attribute)
	return response.Success(c, "Attribute updated successfully", res, nil)
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

func mapToAttributeRes(attr *domain.Attribute) dto.AttributeRes {
	opts := make([]dto.AttributeOptionRes, 0)
	for _, opt := range attr.Options {
		opts = append(opts, dto.AttributeOptionRes{
			ID:    opt.ID,
			Value: opt.Value,
		})
	}
	return dto.AttributeRes{
		ID:        attr.ID,
		Name:      attr.Name,
		Code:      attr.Code,
		ValueType: attr.ValueType,
		Options:   opts,
		CreatedAt: attr.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: attr.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
