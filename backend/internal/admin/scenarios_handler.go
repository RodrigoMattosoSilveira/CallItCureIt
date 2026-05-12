package admin

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

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
	router.Get("/scenarios/:scenarioId", h.GetScenario)
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

func (h *ScenarioHandler) GetScenario(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	scenario, err := h.service.GetScenario(c.Context(), scenarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get admin scenario",
		})
	}

	lines, err := h.service.GetTranscriptWithOpportunities(c.Context(), scenarioID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get admin scenario transcript",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"scenario": mapScenarioDetail(*scenario),
			"lines":    mapScenarioLines(lines),
		},
	})
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

func mapScenarioLines(lines []db.ScenarioLine) []fiber.Map {
	data := make([]fiber.Map, 0, len(lines))

	for _, line := range lines {
		opportunities := make([]fiber.Map, 0, len(line.Opportunities))

		for _, opportunity := range line.Opportunities {
			ruleRefs := make([]fiber.Map, 0, len(opportunity.RuleRefs))

			for _, ruleRef := range opportunity.RuleRefs {
				ruleRefs = append(ruleRefs, fiber.Map{
					"id":           ruleRef.ID,
					"jurisdiction": ruleRef.Jurisdiction,
					"ruleCode":     ruleRef.RuleCode,
					"title":        ruleRef.Title,
					"summary":      ruleRef.Summary,
					"citation":     ruleRef.Citation,
				})
			}

			opportunities = append(opportunities, fiber.Map{
				"id":               opportunity.ID,
				"scenarioLineId":   opportunity.ScenarioLineID,
				"objectionTypeId":  opportunity.ObjectionTypeID,
				"strength":         opportunity.Strength,
				"timingWindow":     opportunity.TimingWindow,
				"explanation":      opportunity.Explanation,
				"expectedPhrase":   opportunity.ExpectedPhrase,
				"isPrimary":        opportunity.IsPrimary,
				"objectionType": fiber.Map{
					"id":            opportunity.ObjectionType.ID,
					"code":          opportunity.ObjectionType.Code,
					"name":          opportunity.ObjectionType.Name,
					"description":   opportunity.ObjectionType.Description,
					"defaultPhrase": opportunity.ObjectionType.DefaultPhrase,
				},
				"ruleRefs": ruleRefs,
			})
		}

		data = append(data, fiber.Map{
			"id":            line.ID,
			"scenarioId":    line.ScenarioID,
			"sequenceNo":    line.SequenceNo,
			"speakerType":   line.SpeakerType,
			"speakerName":   line.SpeakerName,
			"lineText":      line.LineText,
			"lineKind":      line.LineKind,
			"opportunities": opportunities,
		})
	}

	return data
}