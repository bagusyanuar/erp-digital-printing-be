package http

import (
	"math"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/category/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/category/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	categoryUsecase domain.CategoryUsecase
}

func NewCategoryHandler(categoryUsecase domain.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{categoryUsecase: categoryUsecase}
}

func (h *CategoryHandler) Create(c fiber.Ctx) error {
	var req dto.CreateCategoryReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	category := &domain.Category{
		Name: req.Name,
	}

	if err := h.categoryUsecase.Create(c.Context(), category); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create category", err.Error())
	}

	return response.Created(c, "Category created successfully", category)
}

func (h *CategoryHandler) FindAll(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")

	categories, total, err := h.categoryUsecase.FindAll(c.Context(), params, search)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch categories", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Categories fetched successfully", categories, meta)
}

func (h *CategoryHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid category ID", err.Error())
	}

	category, err := h.categoryUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Category not found", err.Error())
	}

	return response.Success(c, "Category fetched successfully", category, nil)
}

func (h *CategoryHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid category ID", err.Error())
	}

	var req dto.UpdateCategoryReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	category, err := h.categoryUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Category not found", err.Error())
	}

	category.Name = req.Name

	if err := h.categoryUsecase.Update(c.Context(), category); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update category", err.Error())
	}

	return response.Success(c, "Category updated successfully", category, nil)
}

func (h *CategoryHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid category ID", err.Error())
	}

	if err := h.categoryUsecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete category", err.Error())
	}

	return response.Success[any](c, "Category deleted successfully", nil, nil)
}
