package auth

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"

	"CallItCureIt/backend/internal/db"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/auth")

	api.Post("/login", h.Login)
	api.Get("/me", h.Me)
	api.Post("/logout", h.Logout)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req loginRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := h.service.Login(c.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserDisabled) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid email or password",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to login",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user":  mapUser(*result.User),
			"token": result.Token,
		},
	})
}

func (h *Handler) Me(c fiber.Ctx) error {
	token := ExtractBearerToken(c.Get("Authorization"))

	user, err := h.service.AuthenticateToken(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapUser(*user),
	})
}

func (h *Handler) Logout(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"ok": true,
		},
	})
}

func mapUser(user db.User) fiber.Map {
	return fiber.Map{
		"id":       user.ID,
		"email":    user.Email,
		"fullName": user.FullName,
		"role":     user.Role,
		"status":   user.Status,
	}
}