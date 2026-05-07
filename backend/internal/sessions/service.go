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