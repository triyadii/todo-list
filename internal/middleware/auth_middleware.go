package middleware

import (
	"strings"

	"todo-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Missing Authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid token format, expected 'Bearer <token>'", nil)
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: "+err.Error(), nil)
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_timezone", claims.Timezone)

		return c.Next()
	}
}
