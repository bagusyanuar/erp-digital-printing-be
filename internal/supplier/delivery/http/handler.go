package http

import (
	"math"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SupplierHandler struct {
	supplierUsecase domain.SupplierUsecase
}

func NewSupplierHandler(supplierUsecase domain.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{supplierUsecase: supplierUsecase}
}

func (h *SupplierHandler) Create(c fiber.Ctx) error {
	var req dto.CreateSupplierReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	supplier := &domain.Supplier{
		Name:        req.Name,
		ContactName: req.ContactName,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
	}

	if err := h.supplierUsecase.Create(c.Context(), supplier); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create supplier", err.Error())
	}

	return response.Created(c, "Supplier created successfully", supplier)
}

func (h *SupplierHandler) FindAll(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	search := c.Query("search", "")

	suppliers, total, err := h.supplierUsecase.FindAll(c.Context(), params, search)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch suppliers", err.Error())
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Suppliers fetched successfully", suppliers, meta)
}

func (h *SupplierHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid supplier ID", err.Error())
	}

	supplier, err := h.supplierUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Supplier not found", err.Error())
	}

	return response.Success(c, "Supplier fetched successfully", supplier, nil)
}

func (h *SupplierHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid supplier ID", err.Error())
	}

	var req dto.UpdateSupplierReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	supplier, err := h.supplierUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Supplier not found", err.Error())
	}

	supplier.Name = req.Name
	supplier.ContactName = req.ContactName
	supplier.Phone = req.Phone
	supplier.Email = req.Email
	supplier.Address = req.Address

	if err := h.supplierUsecase.Update(c.Context(), supplier); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update supplier", err.Error())
	}

	return response.Success(c, "Supplier updated successfully", supplier, nil)
}

func (h *SupplierHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid supplier ID", err.Error())
	}

	if err := h.supplierUsecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete supplier", err.Error())
	}

	return response.Success[any](c, "Supplier deleted successfully", nil, nil)
}
