package scenarios

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

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
	api := app.Group("/api/v1")

	api.Get("/scenarios", h.ListScenarios)
	api.Get("/scenarios/:scenarioId", h.GetScenario)
	api.Get("/scenarios/:scenarioId/transcript", h.GetScenarioTranscript)
}

func (h *Handler) ListScenarios(c fiber.Ctx) error {
	scenarios, err := h.service.ListPublished(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list scenarios",
		})
	}

	data := make([]fiber.Map, 0, len(scenarios))
	for _, scenario := range scenarios {
		data = append(data, mapScenarioSummary(scenario))
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}

func (h *Handler) GetScenario(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	scenario, err := h.service.GetScenario(c.Context(), scenarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get scenario",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapScenarioDetail(*scenario),
	})
}

func (h *Handler) GetScenarioTranscript(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	lines, err := h.service.GetTranscript(c.Context(), scenarioID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get scenario transcript",
		})
	}

	data := make([]fiber.Map, 0, len(lines))
	for _, line := range lines {
		data = append(data, mapScenarioLine(line))
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

func mapScenarioDetail(s db.Scenario) fiber.Map {
	actors := make([]fiber.Map, 0, len(s.Actors))
	for _, actor := range s.Actors {
		actors = append(actors, fiber.Map{
			"id":        actor.ID,
			"name":      actor.Name,
			"actorType": actor.ActorType,
			"persona":   actor.Persona,
		})
	}

	summary := mapScenarioSummary(s)
	summary["actors"] = actors

	return summary
}

func mapScenarioLine(line db.ScenarioLine) fiber.Map {
	return fiber.Map{
		"id":          line.ID,
		"scenarioId":  line.ScenarioID,
		"sequenceNo":  line.SequenceNo,
		"speakerType": line.SpeakerType,
		"speakerName": line.SpeakerName,
		"lineText":    line.LineText,
		"lineKind":    line.LineKind,
	}
}
