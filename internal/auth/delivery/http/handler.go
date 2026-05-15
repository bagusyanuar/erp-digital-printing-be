package http

import (
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
	cfg         *config.Config
}

func NewAuthHandler(authUsecase domain.AuthUsecase, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		cfg:         cfg,
	}
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
	h.setRefreshTokenCookie(c, refreshToken)

	return response.Success(c, "Login successful", dto.LoginRes{
		AccessToken: accessToken,
	}, nil)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	// 1. Get Refresh Token from Cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return response.Error(c, fiber.StatusUnauthorized, "Refresh token missing", nil)
	}

	// 2. Call Usecase
	newAccessToken, newRefreshToken, err := h.authUsecase.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err.Error(), nil)
	}

	// 3. Set New Refresh Token in Cookie (Rotation)
	h.setRefreshTokenCookie(c, newRefreshToken)

	return response.Success(c, "Token refreshed successfully", dto.LoginRes{
		AccessToken: newAccessToken,
	}, nil)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "None",
		Path:     "/",
	})

	return response.Success[any](c, "Logout successful", nil, nil)
}

func (h *AuthHandler) setRefreshTokenCookie(c fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HTTPOnly: true,
		Secure:   true,   // Wajib true kalau SameSite=None
		SameSite: "None", // Biar cookie nggak hilang pas reload di cross-site context
		Path:     "/",
	})
}
