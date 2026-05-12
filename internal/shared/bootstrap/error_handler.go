package bootstrap

import (
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func NewGlobalErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		log.Error("Request error",
			zap.Int("status", code),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Error(err),
		)

		return response.Error(c, code, message, nil)
	}
}
