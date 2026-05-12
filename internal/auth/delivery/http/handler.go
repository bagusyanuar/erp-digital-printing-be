package http

import (
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	accessToken, refreshToken, err := h.authUsecase.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err.Error(), nil)
	}

	// Set Refresh Token in Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 30), // Standard 30 days
		HTTPOnly: true,
		Secure:   false, // Should be true in production with HTTPS
		SameSite: "Lax",
		Path:     "/",
	})

	return response.Success(c, "Login successful", dto.LoginRes{
		AccessToken: accessToken,
	}, nil)
}
