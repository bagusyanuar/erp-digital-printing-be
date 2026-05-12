package http

import (
	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUserHandler(userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

func (h *UserHandler) Create(c fiber.Ctx) error {
	var req dto.CreateUserReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	user := &domain.User{
		Username: req.Username,
		Password: req.Password,
	}

	if err := h.userUsecase.Create(c.Context(), user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create user", err.Error())
	}

	return response.Created(c, "User created successfully", user)
}

func (h *UserHandler) FindAll(c fiber.Ctx) error {
	users, err := h.userUsecase.FindAll(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}

	return response.Success(c, "Users fetched successfully", users, nil)
}

func (h *UserHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	user, err := h.userUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	return response.Success(c, "User fetched successfully", user, nil)
}

func (h *UserHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	var req dto.UpdateUserReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	user, err := h.userUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	user.Username = req.Username
	if req.Password != "" {
		user.Password = req.Password
	}

	if err := h.userUsecase.Update(c.Context(), user); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update user", err.Error())
	}

	return response.Success(c, "User updated successfully", user, nil)
}

func (h *UserHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	if err := h.userUsecase.Delete(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}

	return response.Success[any](c, "User deleted successfully", nil, nil)
}
