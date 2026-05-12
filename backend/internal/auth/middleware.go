package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

const ContextUserID = "userID"
const ContextUserEmail = "userEmail"
const ContextUserRole = "userRole"

func RequireAuth(service *Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header is required",
			})
		}

		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "bearer token is required",
			})
		}

		claims, err := service.ParseToken(strings.TrimSpace(tokenString))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals(ContextUserID, claims.UserID)
		c.Locals(ContextUserEmail, claims.Email)
		c.Locals(ContextUserRole, claims.Role)

		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals(ContextUserRole).(string)

		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin role is required",
			})
		}

		return c.Next()
	}
}