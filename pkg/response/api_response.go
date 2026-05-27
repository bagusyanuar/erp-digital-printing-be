package response

import (
	"reflect"

	"github.com/gofiber/fiber/v3"
)

type APIResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Pagination struct {
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	CurrentPage int   `json:"current_page"`
	Limit       int   `json:"limit"`
}

type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Success response helper
func Success[T any](c fiber.Ctx, message string, data T, meta any) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse[any]{
		Message: message,
		Data:    ensureEmptySlice(data),
		Meta:    meta,
	})
}

// Error response helper
func Error(c fiber.Ctx, status int, message string, errors any) error {
	return c.Status(status).JSON(APIResponse[any]{
		Message: message,
		Errors:  errors,
	})
}

// Created response helper
func Created[T any](c fiber.Ctx, message string, data T) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse[any]{
		Message: message,
		Data:    ensureEmptySlice(data),
	})
}

func ensureEmptySlice(data any) any {
	if data == nil {
		return nil
	}
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return nil
	}
	if val.Kind() == reflect.Slice {
		if val.IsNil() || val.Len() == 0 {
			return []any{}
		}
	}
	return data
}
