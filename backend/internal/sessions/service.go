package sessions

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/objections"
)

var ErrNoMoreLines = errors.New("no more lines")
var ErrSessionCompleted = errors.New("session already completed")
var ErrScenarioNotFound = errors.New("scenario not found")
var ErrInvalidAction = errors.New("invalid trainee action")
var ErrNoCurrentLine = errors.New("session has no current line")

type CreateSessionInput struct {
	ScenarioID string
	Mode       string
	UserID     string
}

type AdvanceSessionResult struct {
	Session   *db.Session
	Line      *db.ScenarioLine
	Completed bool
}

type SubmitActionInput struct {
	SessionID  string
	ActionType string
	RawText   string
}

type SubmitActionResult struct {
	Session    *db.Session
	Action     *db.TraineeAction
	Event      *db.SessionEvent
	Evaluation *db.ActionEvaluation
}

type Service struct {
	repo Repository
	objectionMatcher *objections.Matcher
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:             repo,
		objectionMatcher: objections.NewMatcher(),
	}
}

func (s *Service) CreateSession(ctx context.Context, input CreateSessionInput) (*db.Session, error) {
	if input.Mode == "" {
		input.Mode = "spot_objection"
	}

	lineCount, err := s.repo.CountScenarioLines(ctx, input.ScenarioID)
	if err != nil {
		return nil, err
	}

	if lineCount == 0 {
		return nil, ErrScenarioNotFound
	}

	session := &db.Session{
		ID:                uuid.NewString(),
		UserID:            input.UserID,
		ScenarioID:        input.ScenarioID,
		Status:            "active",
		CurrentSequenceNo: 0,
		Mode:              input.Mode,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*db.Session, []db.SessionEvent, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	events, err := s.repo.ListEvents(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	return session, events, nil
}

func (s *Service) AdvanceSession(ctx context.Context, sessionID string) (*AdvanceSessionResult, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status == "completed" {
		return nil, ErrSessionCompleted
	}

	line, err := s.repo.GetNextScenarioLine(
		ctx,
		session.ScenarioID,
		session.CurrentSequenceNo,
	)

	if errors.Is(err, ErrNoMoreLines) {
		now := time.Now()
		session.Status = "completed"
		session.CompletedAt = &now

		if updateErr := s.repo.Update(ctx, session); updateErr != nil {
			return nil, updateErr
		}

		return &AdvanceSessionResult{
			Session:   session,
			Line:      nil,
			Completed: true,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	event := &db.SessionEvent{
		ID:         uuid.NewString(),
		SessionID:  session.ID,
		SequenceNo: line.SequenceNo,
		EventType:  "system_line",
		Actor:      line.SpeakerName,
		Text:       line.LineText,
	}

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return nil, err
	}

	session.CurrentSequenceNo = line.SequenceNo

	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}

	return &AdvanceSessionResult{
		Session:   session,
		Line:      line,
		Completed: false,
	}, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (s *Service) SubmitAction(
	ctx context.Context,
	input SubmitActionInput,
) (*SubmitActionResult, error) {
	if input.ActionType == "" {
		input.ActionType = "object"
	}

	if input.ActionType != "object" &&
		input.ActionType != "respond" &&
		input.ActionType != "pass" {
		return nil, ErrInvalidAction
	}

	if input.ActionType != "pass" && input.RawText == "" {
		return nil, ErrInvalidAction
	}

	session, err := s.repo.GetByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}

	if session.Status == "completed" {
		return nil, ErrSessionCompleted
	}

	if session.CurrentSequenceNo <= 0 {
		return nil, ErrNoCurrentLine
	}

	line, err := s.repo.GetScenarioLineWithOpportunities(
		ctx,
		session.ScenarioID,
		session.CurrentSequenceNo,
	)
	if err != nil {
		return nil, err
	}

	rawText := input.RawText
	if input.ActionType == "pass" && rawText == "" {
		rawText = "Pass"
	}

	var normalizedObjectionTypeID *string

	normalized := s.objectionMatcher.Normalize(rawText)
	if normalized.Matched {
		id := normalized.ObjectionTypeID
		normalizedObjectionTypeID = &id
	}

	scenarioLineID := line.ID

	action := &db.TraineeAction{
		ID:                        uuid.NewString(),
		SessionID:                 session.ID,
		ScenarioLineID:            &scenarioLineID,
		ActionType:                input.ActionType,
		RawText:                   rawText,
		NormalizedObjectionTypeID: normalizedObjectionTypeID,
	}

	if err := s.repo.CreateTraineeAction(ctx, action); err != nil {
		return nil, err
	}

	evaluation := s.evaluateAction(action, line, normalized)

	if err := s.repo.CreateActionEvaluation(ctx, evaluation); err != nil {
		return nil, err
	}

	eventType := "trainee_objection"
	if input.ActionType == "respond" {
		eventType = "trainee_response"
	}
	if input.ActionType == "pass" {
		eventType = "coach_feedback"
	}

	event := &db.SessionEvent{
		ID:         uuid.NewString(),
		SessionID:  session.ID,
		SequenceNo: session.CurrentSequenceNo,
		EventType:  eventType,
		Actor:      "Trainee Counsel",
		Text:       rawText,
	}

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return nil, err
	}

	feedbackEvent := &db.SessionEvent{
		ID:         uuid.NewString(),
		SessionID:  session.ID,
		SequenceNo: session.CurrentSequenceNo,
		EventType:  "coach_feedback",
		Actor:      "Coach",
		Text:       evaluation.Feedback,
	}

	if err := s.repo.CreateEvent(ctx, feedbackEvent); err != nil {
		return nil, err
	}

	return &SubmitActionResult{
		Session:    session,
		Action:     action,
		Event:      event,
		Evaluation: evaluation,
	}, nil
}
// For now, a pass is stored as a simple event so we can support “I do not 
// object” later.  In the deterministic matching phase, pass will be used to 
// detect whether the trainee missed an objection opportunity.

func (s *Service) evaluateAction(
	action *db.TraineeAction,
	line *db.ScenarioLine,
	normalized objections.NormalizedObjection,
) *db.ActionEvaluation {
	evaluation := &db.ActionEvaluation{
		ID:                uuid.NewString(),
		TraineeActionID:   action.ID,
		Valid:             false,
		Timely:            false,
		Ruling:            "overruled",
		LegalAccuracyScore: 0,
		PhrasingScore:     scorePhrasing(action.RawText),
		StrategyScore:     0,
		Feedback:          "That objection does not match a recognized objection opportunity on this line.",
	}

	if action.NormalizedObjectionTypeID != nil {
		evaluation.NormalizedObjectionTypeID = action.NormalizedObjectionTypeID
	}

	if action.ActionType == "pass" {
		if len(line.Opportunities) > 0 {
			evaluation.Ruling = "no_ruling"
			evaluation.Feedback = "You passed, but this line had at least one objection opportunity."
			evaluation.StrategyScore = 0
			return evaluation
		}

		evaluation.Valid = true
		evaluation.Timely = true
		evaluation.Ruling = "no_ruling"
		evaluation.LegalAccuracyScore = 100
		evaluation.PhrasingScore = 100
		evaluation.StrategyScore = 100
		evaluation.Feedback = "Correct. There was no strong objection opportunity on this line."
		return evaluation
	}

	if !normalized.Matched {
		evaluation.Feedback = "The system could not identify a supported objection type from your text."
		return evaluation
	}

	for _, opportunity := range line.Opportunities {
		if opportunity.ObjectionTypeID == normalized.ObjectionTypeID {
			matchedOpportunityID := opportunity.ID

			evaluation.MatchedOpportunityID = &matchedOpportunityID
			evaluation.Valid = true
			evaluation.Timely = true
			evaluation.Ruling = "sustained"
			evaluation.LegalAccuracyScore = scoreLegalAccuracy(opportunity.Strength)
			evaluation.StrategyScore = scoreStrategy(opportunity.Strength)
			evaluation.Feedback = buildCorrectFeedback(opportunity)

			return evaluation
		}
	}

	if len(line.Opportunities) > 0 {
		expected := line.Opportunities[0]

		evaluation.Feedback = "This line had an objection opportunity, but your objection ground did not match the strongest expected ground. Expected: " +
			expected.ObjectionType.Name + "."
		evaluation.LegalAccuracyScore = 25
		evaluation.StrategyScore = 25

		return evaluation
	}

	evaluation.Feedback = "There was no expected objection opportunity on this line, so the objection would likely be overruled."
	evaluation.LegalAccuracyScore = 0
	evaluation.StrategyScore = 0

	return evaluation
}

func scoreLegalAccuracy(strength string) float64 {
	switch strength {
	case "strong":
		return 100
	case "moderate":
		return 80
	case "weak":
		return 60
	default:
		return 50
	}
}

func scoreStrategy(strength string) float64 {
	switch strength {
	case "strong":
		return 100
	case "moderate":
		return 75
	case "weak":
		return 50
	default:
		return 50
	}
}

func scorePhrasing(rawText string) float64 {
	text := strings.TrimSpace(rawText)

	if text == "" {
		return 0
	}

	score := 70.0

	lower := strings.ToLower(text)

	if strings.Contains(lower, "objection") {
		score += 15
	}

	if len(text) <= 80 {
		score += 15
	}

	if score > 100 {
		return 100
	}

	return score
}

func buildCorrectFeedback(opportunity db.ObjectionOpportunity) string {
	if opportunity.Explanation != "" {
		return "Correct. " + opportunity.Explanation
	}

	return "Correct. The objection matches the expected ground: " + opportunity.ObjectionType.Name + "."
}