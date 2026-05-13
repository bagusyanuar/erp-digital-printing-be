package middleware

import (
	"strings"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/jwt"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(jwtUtil jwt.JWTUtil) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "Missing authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid authorization header format", nil)
		}

		tokenString := parts[1]
		claims, err := jwtUtil.ParseToken(tokenString)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid or expired token", err.Error())
		}

		// Store user info in locals
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)

		return c.Next()
	}
}
