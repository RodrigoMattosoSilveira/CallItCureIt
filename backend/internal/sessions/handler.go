package sessions

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
	api := app.Group("/api/v1")

	api.Post("/sessions", h.CreateSession)
	api.Get("/sessions/:sessionId", h.GetSession)
	api.Post("/sessions/:sessionId/next", h.NextLine)
	api.Post("/sessions/:sessionId/actions", h.SubmitAction)
}

type createSessionRequest struct {
	ScenarioID string `json:"scenarioId"`
	Mode       string `json:"mode"`
}

type submitActionRequest struct {
	ActionType string `json:"actionType"`
	RawText    string `json:"rawText"`
}

func (h *Handler) CreateSession(c fiber.Ctx) error {
	var req createSessionRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.ScenarioID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "scenarioId is required",
		})
	}

	session, err := h.service.CreateSession(c.Context(), CreateSessionInput{
		ScenarioID: req.ScenarioID,
		Mode:       req.Mode,
	})

	if err != nil {
		if errors.Is(err, ErrScenarioNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "scenario not found or has no lines",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": mapSession(*session, nil),
	})
}

func (h *Handler) GetSession(c fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	session, events, err := h.service.GetSession(c.Context(), sessionID)
	if err != nil {
		if IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get session",
		})
	}

	return c.JSON(fiber.Map{
		"data": mapSession(*session, events),
	})
}

func (h *Handler) NextLine(c fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	result, err := h.service.AdvanceSession(c.Context(), sessionID)
	if err != nil {
		if IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session not found",
			})
		}

		if errors.Is(err, ErrSessionCompleted) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "session is already completed",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to advance session",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"session":   mapSessionSummary(*result.Session),
			"line":      mapScenarioLinePtr(result.Line),
			"completed": result.Completed,
		},
	})
}

func mapSession(session db.Session, events []db.SessionEvent) fiber.Map {
	return fiber.Map{
		"id":                session.ID,
		"scenarioId":        session.ScenarioID,
		"status":            session.Status,
		"currentSequenceNo": session.CurrentSequenceNo,
		"mode":              session.Mode,
		"startedAt":         session.StartedAt,
		"completedAt":       session.CompletedAt,
		"events":            mapSessionEvents(events),
	}
}

func mapSessionSummary(session db.Session) fiber.Map {
	return fiber.Map{
		"id":                session.ID,
		"scenarioId":        session.ScenarioID,
		"status":            session.Status,
		"currentSequenceNo": session.CurrentSequenceNo,
		"mode":              session.Mode,
	}
}

func mapSessionEvents(events []db.SessionEvent) []fiber.Map {
	data := make([]fiber.Map, 0, len(events))

	for _, event := range events {
		data = append(data, fiber.Map{
			"id":         event.ID,
			"sessionId":  event.SessionID,
			"sequenceNo": event.SequenceNo,
			"eventType":  event.EventType,
			"actor":      event.Actor,
			"text":       event.Text,
			"createdAt":  event.CreatedAt,
		})
	}

	return data
}

func mapScenarioLinePtr(line *db.ScenarioLine) fiber.Map {
	if line == nil {
		return nil
	}

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

func (h *Handler) SubmitAction(c fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req submitActionRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := h.service.SubmitAction(c.Context(), SubmitActionInput{
		SessionID:  sessionID,
		ActionType: req.ActionType,
		RawText:    req.RawText,
	})

	if err != nil {
		if IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "session or current line not found",
			})
		}

		if errors.Is(err, ErrSessionCompleted) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "session is already completed",
			})
		}

		if errors.Is(err, ErrNoCurrentLine) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "advance the session before submitting an action",
			})
		}

		if errors.Is(err, ErrInvalidAction) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid action",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to submit action",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"session":    mapSessionSummary(*result.Session),
			"action":     mapTraineeAction(*result.Action),
			"event":      mapSessionEvent(*result.Event),
			"evaluation": mapActionEvaluation(*result.Evaluation),
		},
	})
}

func mapTraineeAction(action db.TraineeAction) fiber.Map {
	return fiber.Map{
		"id":             action.ID,
		"sessionId":      action.SessionID,
		"scenarioLineId": action.ScenarioLineID,
		"actionType":     action.ActionType,
		"rawText":        action.RawText,
		"createdAt":      action.CreatedAt,
	}
}

func mapSessionEvent(event db.SessionEvent) fiber.Map {
	return fiber.Map{
		"id":         event.ID,
		"sessionId":  event.SessionID,
		"sequenceNo": event.SequenceNo,
		"eventType":  event.EventType,
		"actor":      event.Actor,
		"text":       event.Text,
		"createdAt":  event.CreatedAt,
	}
}

func mapActionEvaluation(evaluation db.ActionEvaluation) fiber.Map {
	return fiber.Map{
		"id":                         evaluation.ID,
		"traineeActionId":            evaluation.TraineeActionID,
		"matchedOpportunityId":       evaluation.MatchedOpportunityID,
		"normalizedObjectionTypeId":  evaluation.NormalizedObjectionTypeID,
		"valid":                      evaluation.Valid,
		"timely":                     evaluation.Timely,
		"ruling":                     evaluation.Ruling,
		"legalAccuracyScore":         evaluation.LegalAccuracyScore,
		"phrasingScore":              evaluation.PhrasingScore,
		"strategyScore":              evaluation.StrategyScore,
		"feedback":                   evaluation.Feedback,
		"createdAt":                  evaluation.CreatedAt,
	}
}