package auth

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"
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
	api := app.Group("/api/v1")

	api.Post("/auth/login", h.Login)
	api.Get("/auth/me", RequireAuth(h.service), h.Me)
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
		if errors.Is(err, ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid email or password",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to log in",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"token": result.Token,
			"user": fiber.Map{
				"id":    result.User.ID,
				"email": result.User.Email,
				"name":  result.User.Name,
				"role":  result.User.Role,
			},
		},
	})
}

func (h *Handler) Me(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":    c.Locals(ContextUserID),
			"email": c.Locals(ContextUserEmail),
			"role":  c.Locals(ContextUserRole),
		},
	})
}