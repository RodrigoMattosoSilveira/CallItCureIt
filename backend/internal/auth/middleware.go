package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"CallItCureIt/backend/internal/db"
)

const userContextKey = "auth.user"

func RequireAuth(service *Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := ExtractBearerToken(c.Get("Authorization"))

		user, err := service.AuthenticateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		c.Locals(userContextKey, user)

		return c.Next()
	}
}

func RequireAdmin(service *Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := ExtractBearerToken(c.Get("Authorization"))

		user, err := service.AuthenticateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
			})
		}

		c.Locals(userContextKey, user)

		return c.Next()
	}
}

func CurrentUser(c fiber.Ctx) (*db.User, bool) {
	value := c.Locals(userContextKey)

	user, ok := value.(*db.User)
	if !ok {
		return nil, false
	}

	return user, true
}

func ExtractBearerToken(header string) string {
	header = strings.TrimSpace(header)

	if header == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}