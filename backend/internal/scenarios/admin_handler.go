package scenarios

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
)

type AdminHandler struct {
	service *AdminService
}

func NewAdminHandler(service *AdminService) *AdminHandler {
	return &AdminHandler{
		service: service,
	}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App, middleware ...fiber.Handler) {
	api := app.Group("/api/v1/admin")

	for _, handler := range middleware {
		api.Use(handler)
	}

	api.Get("/scenarios", h.ListScenarios)
	api.Post("/scenarios", h.CreateScenario)
	api.Get("/scenarios/:scenarioId", h.GetScenario)
	api.Put("/scenarios/:scenarioId", h.UpdateScenario)
	api.Post("/scenarios/:scenarioId/publish", h.PublishScenario)
	api.Post("/scenarios/:scenarioId/archive", h.ArchiveScenario)

	api.Post("/scenarios/:scenarioId/lines", h.CreateScenarioLine)

	api.Put("/scenario-lines/:lineId", h.UpdateScenarioLine)
	api.Delete("/scenario-lines/:lineId", h.DeleteScenarioLine)

	api.Get("/objection-types", h.ListObjectionTypes)
	api.Post("/scenario-lines/:lineId/opportunities", h.CreateOpportunity)

	api.Put("/opportunities/:opportunityId", h.UpdateOpportunity)
	api.Delete("/opportunities/:opportunityId", h.DeleteOpportunity)
}

type createScenarioRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Jurisdiction string `json:"jurisdiction"`
	PracticeArea string `json:"practiceArea"`
	HearingType  string `json:"hearingType"`
	Difficulty   string `json:"difficulty"`
	Status       string `json:"status"`
}

type updateScenarioRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Jurisdiction string `json:"jurisdiction"`
	PracticeArea string `json:"practiceArea"`
	HearingType  string `json:"hearingType"`
	Difficulty   string `json:"difficulty"`
	Status       string `json:"status"`
}

type createScenarioLineRequest struct {
	SequenceNo  int    `json:"sequenceNo"`
	SpeakerType string `json:"speakerType"`
	SpeakerName string `json:"speakerName"`
	LineText    string `json:"lineText"`
	LineKind    string `json:"lineKind"`
}

type createOpportunityRequest struct {
	ObjectionTypeID string `json:"objectionTypeId"`
	Strength        string `json:"strength"`
	TimingWindow    string `json:"timingWindow"`
	Explanation     string `json:"explanation"`
	ExpectedPhrase  string `json:"expectedPhrase"`
	IsPrimary       bool   `json:"isPrimary"`
}

type updateScenarioLineRequest struct {
	SequenceNo  int    `json:"sequenceNo"`
	SpeakerType string `json:"speakerType"`
	SpeakerName string `json:"speakerName"`
	LineText    string `json:"lineText"`
	LineKind    string `json:"lineKind"`
}

type updateOpportunityRequest struct {
	ObjectionTypeID string `json:"objectionTypeId"`
	Strength        string `json:"strength"`
	TimingWindow    string `json:"timingWindow"`
	Explanation     string `json:"explanation"`
	ExpectedPhrase  string `json:"expectedPhrase"`
	IsPrimary       bool   `json:"isPrimary"`
}

func (h *AdminHandler) ListScenarios(c fiber.Ctx) error {
	scenarios, err := h.service.ListScenarios(c.Context())
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

func (h *AdminHandler) CreateScenario(c fiber.Ctx) error {
	var req createScenarioRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	scenario, err := h.service.CreateScenario(c.Context(), CreateScenarioInput{
		Title:        req.Title,
		Description:  req.Description,
		Jurisdiction: req.Jurisdiction,
		PracticeArea: req.PracticeArea,
		HearingType:  req.HearingType,
		Difficulty:   req.Difficulty,
		Status:       req.Status,
	})

	if err != nil {
		if errors.Is(err, ErrInvalidScenario) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid scenario",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create scenario",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": mapScenarioSummary(*scenario),
	})
}

func (h *AdminHandler) GetScenario(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	scenario, lines, err := h.service.GetScenario(c.Context(), scenarioID)
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
		"data": fiber.Map{
			"scenario": mapScenarioDetail(*scenario),
			"lines":    mapAdminScenarioLines(lines),
		},
	})
}

func (h *AdminHandler) UpdateScenario(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	var req updateScenarioRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	scenario, err := h.service.UpdateScenario(c.Context(), scenarioID, UpdateScenarioInput{
		Title:        req.Title,
		Description:  req.Description,
		Jurisdiction: req.Jurisdiction,
		PracticeArea: req.PracticeArea,
		HearingType:  req.HearingType,
		Difficulty:   req.Difficulty,
		Status:       req.Status,
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update scenario",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapScenarioSummary(*scenario),
	})
}

func (h *AdminHandler) PublishScenario(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	scenario, err := h.service.PublishScenario(c.Context(), scenarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to publish scenario",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapScenarioSummary(*scenario),
	})
}

func (h *AdminHandler) ArchiveScenario(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	scenario, err := h.service.ArchiveScenario(c.Context(), scenarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to archive scenario",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapScenarioSummary(*scenario),
	})
}

func (h *AdminHandler) CreateScenarioLine(c fiber.Ctx) error {
	scenarioID := c.Params("scenarioId")

	var req createScenarioLineRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	line, err := h.service.CreateScenarioLine(c.Context(), CreateScenarioLineInput{
		ScenarioID:   scenarioID,
		SequenceNo:   req.SequenceNo,
		SpeakerType:  req.SpeakerType,
		SpeakerName:  req.SpeakerName,
		LineText:     req.LineText,
		LineKind:     req.LineKind,
	})

	if err != nil {
		if errors.Is(err, ErrInvalidScenarioLine) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid scenario line",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create scenario line",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": mapScenarioLine(*line),
	})
}

func (h *AdminHandler) ListObjectionTypes(c fiber.Ctx) error {
	objectionTypes, err := h.service.ListObjectionTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list objection types",
		})
	}

	data := make([]fiber.Map, 0, len(objectionTypes))
	for _, objectionType := range objectionTypes {
		data = append(data, mapObjectionType(objectionType))
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}

func (h *AdminHandler) CreateOpportunity(c fiber.Ctx) error {
	lineID := c.Params("lineId")

	var req createOpportunityRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	opportunity, err := h.service.CreateOpportunity(c.Context(), CreateOpportunityInput{
		ScenarioLineID:  lineID,
		ObjectionTypeID: req.ObjectionTypeID,
		Strength:        req.Strength,
		TimingWindow:    req.TimingWindow,
		Explanation:     req.Explanation,
		ExpectedPhrase:  req.ExpectedPhrase,
		IsPrimary:       req.IsPrimary,
	})

	if err != nil {
		if errors.Is(err, ErrInvalidOpportunity) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid objection opportunity",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create objection opportunity",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": mapObjectionOpportunity(*opportunity),
	})
}

func mapAdminScenarioLines(lines []db.ScenarioLine) []fiber.Map {
	data := make([]fiber.Map, 0, len(lines))

	for _, line := range lines {
		item := mapScenarioLine(line)
		opportunities := make([]fiber.Map, 0, len(line.Opportunities))

		for _, opportunity := range line.Opportunities {
			opportunities = append(opportunities, mapObjectionOpportunity(opportunity))
		}

		item["opportunities"] = opportunities
		data = append(data, item)
	}

	return data
}

func mapObjectionType(objectionType db.ObjectionType) fiber.Map {
	return fiber.Map{
		"id":            objectionType.ID,
		"code":          objectionType.Code,
		"name":          objectionType.Name,
		"description":   objectionType.Description,
		"defaultPhrase": objectionType.DefaultPhrase,
	}
}

func mapObjectionOpportunity(opportunity db.ObjectionOpportunity) fiber.Map {
	return fiber.Map{
		"id":              opportunity.ID,
		"scenarioLineId":  opportunity.ScenarioLineID,
		"objectionTypeId": opportunity.ObjectionTypeID,
		"strength":        opportunity.Strength,
		"timingWindow":    opportunity.TimingWindow,
		"explanation":     opportunity.Explanation,
		"expectedPhrase":  opportunity.ExpectedPhrase,
		"isPrimary":       opportunity.IsPrimary,
	}
}

func (h *AdminHandler) UpdateScenarioLine(c fiber.Ctx) error {
	lineID := c.Params("lineId")

	var req updateScenarioLineRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	line, err := h.service.UpdateScenarioLine(c.Context(), lineID, UpdateScenarioLineInput{
		SequenceNo:  req.SequenceNo,
		SpeakerType: req.SpeakerType,
		SpeakerName: req.SpeakerName,
		LineText:    req.LineText,
		LineKind:    req.LineKind,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario line not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update scenario line",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapScenarioLine(*line),
	})
}

func (h *AdminHandler) DeleteScenarioLine(c fiber.Ctx) error {
	lineID := c.Params("lineId")

	if err := h.service.DeleteScenarioLine(c.Context(), lineID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete scenario line",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) UpdateOpportunity(c fiber.Ctx) error {
	opportunityID := c.Params("opportunityId")

	var req updateOpportunityRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	opportunity, err := h.service.UpdateOpportunity(c.Context(), opportunityID, UpdateOpportunityInput{
		ObjectionTypeID: req.ObjectionTypeID,
		Strength:        req.Strength,
		TimingWindow:    req.TimingWindow,
		Explanation:     req.Explanation,
		ExpectedPhrase:  req.ExpectedPhrase,
		IsPrimary:       req.IsPrimary,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "objection opportunity not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update objection opportunity",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapObjectionOpportunity(*opportunity),
	})
}

func (h *AdminHandler) DeleteOpportunity(c fiber.Ctx) error {
	opportunityID := c.Params("opportunityId")

	if err := h.service.DeleteOpportunity(c.Context(), opportunityID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete objection opportunity",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}