package middleware

import (
	"strings"

	"github.com/bagusyanuar/erp-digital-printing-be/pkg/casbin"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var actionMap = map[string]string{
	"GET":    "read",
	"POST":   "create",
	"PUT":    "update",
	"PATCH":  "update",
	"DELETE": "delete",
}

// extractResource extracts the resource name from the request path.
// Path format: /api/v1/{resource}/... or /api/v1/{group}/{resource}/...
// Examples:
//   - /api/v1/users       -> "users"
//   - /api/v1/users/123   -> "users"
//   - /api/v1/rbac/roles  -> "rbac"
func extractResource(path string) string {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 3 {
		return ""
	}
	return pathParts[2] // api=0, v1=1, resource=2
}

func RBACMiddleware(enforcer *casbin.CasbinHelper) fiber.Handler {
	return func(c fiber.Ctx) error {
		_, ok := c.Locals("user_id").(uuid.UUID)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found in context", nil)
		}

		// TODO: Temporary bypass for development - allowing all roles to access endpoints
		return c.Next()
	}
}
