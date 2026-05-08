package sessions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
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
	Session *db.Session
	Action  *db.TraineeAction
	Event   *db.SessionEvent
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
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

	line, err := s.repo.GetScenarioLineBySequence(
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

	scenarioLineID := line.ID

	action := &db.TraineeAction{
		ID:             uuid.NewString(),
		SessionID:      session.ID,
		ScenarioLineID: &scenarioLineID,
		ActionType:     input.ActionType,
		RawText:        rawText,
	}

	if err := s.repo.CreateTraineeAction(ctx, action); err != nil {
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

	return &SubmitActionResult{
		Session: session,
		Action:  action,
		Event:   event,
	}, nil
}
// For now, a pass is stored as a simple event so we can support “I do not 
// object” later.  In the deterministic matching phase, pass will be used to 
// detect whether the trainee missed an objection opportunity.