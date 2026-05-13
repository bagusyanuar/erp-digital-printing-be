package http

import (
	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type RBACHandler struct {
	usecase domain.RBACUsecase
}

func NewRBACHandler(usecase domain.RBACUsecase) *RBACHandler {
	return &RBACHandler{usecase: usecase}
}

func (h *RBACHandler) CreateRole(c fiber.Ctx) error {
	var req dto.CreateRoleReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.usecase.CreateRole(c.Context(), role); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create role", err.Error())
	}

	return response.Created(c, "Role created successfully", role)
}

func (h *RBACHandler) FindAllRoles(c fiber.Ctx) error {
	roles, err := h.usecase.FindAllRoles(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch roles", err.Error())
	}

	return response.Success(c, "Roles fetched successfully", roles, nil)
}

func (h *RBACHandler) DeleteRole(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid Role ID", err.Error())
	}

	if err := h.usecase.DeleteRole(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete role", err.Error())
	}

	return response.Success[any](c, "Role deleted successfully", nil, nil)
}

func (h *RBACHandler) AssignRoleToUser(c fiber.Ctx) error {
	var req dto.AssignRoleReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.usecase.AssignRoleToUser(c.Context(), req.UserID, req.RoleID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to assign role", err.Error())
	}

	return response.Success[any](c, "Role assigned to user successfully", nil, nil)
}
