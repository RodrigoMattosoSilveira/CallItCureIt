package admin

import (
	"github.com/gofiber/fiber/v3"

	"CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/scenarios"
)

type ScenarioHandler struct {
	service *scenarios.Service
}

func NewScenarioHandler(service *scenarios.Service) *ScenarioHandler {
	return &ScenarioHandler{
		service: service,
	}
}

func (h *ScenarioHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/scenarios", h.ListScenarios)
	router.Get("/objection-types", h.ListObjectionTypes)
}

func (h *ScenarioHandler) ListScenarios(c fiber.Ctx) error {
	items, err := h.service.ListPublished(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list admin scenarios",
		})
	}

	data := make([]fiber.Map, 0, len(items))
	for _, scenario := range items {
		data = append(data, mapScenarioSummary(scenario))
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}

func mapScenarioSummary(s db.Scenario) fiber.Map {
	return fiber.Map{
		"id":           s.ID,
		"title":        s.Title,
		"description":  s.Description,
		"jurisdiction": s.Jurisdiction,
		"practiceArea": s.PracticeArea,
		"hearingType":  s.HearingType,
		"difficulty":   s.Difficulty,
		"status":       s.Status,
	}
}

func (h *ScenarioHandler) ListObjectionTypes(c fiber.Ctx) error {
	items, err := h.service.ListObjectionTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list objection types",
		})
	}

	data := make([]fiber.Map, 0, len(items))
	for _, item := range items {
		data = append(data, fiber.Map{
			"id":            item.ID,
			"code":          item.Code,
			"name":          item.Name,
			"description":   item.Description,
			"defaultPhrase": item.DefaultPhrase,
		})
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}