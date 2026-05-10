package scenarios

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"CallItCureIt/backend/internal/db"
)

var ErrInvalidScenario = errors.New("invalid scenario")
var ErrInvalidScenarioLine = errors.New("invalid scenario line")
var ErrInvalidOpportunity = errors.New("invalid objection opportunity")

type AdminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) *AdminService {
	return &AdminService{
		repo: repo,
	}
}

type CreateScenarioInput struct {
	Title        string
	Description  string
	Jurisdiction string
	PracticeArea string
	HearingType  string
	Difficulty   string
	Status       string
}

type UpdateScenarioInput struct {
	Title        string
	Description  string
	Jurisdiction string
	PracticeArea string
	HearingType  string
	Difficulty   string
	Status       string
}

type CreateScenarioLineInput struct {
	ScenarioID   string
	SequenceNo   int
	SpeakerType  string
	SpeakerName  string
	LineText     string
	LineKind     string
}

type CreateOpportunityInput struct {
	ScenarioLineID   string
	ObjectionTypeID  string
	Strength         string
	TimingWindow     string
	Explanation      string
	ExpectedPhrase   string
	IsPrimary        bool
}

func (s *AdminService) ListScenarios(ctx context.Context) ([]db.Scenario, error) {
	return s.repo.ListAll(ctx)
}

func (s *AdminService) GetScenario(ctx context.Context, id string) (*db.Scenario, []db.ScenarioLine, error) {
	scenario, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	lines, err := s.repo.ListLines(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return scenario, lines, nil
}

func (s *AdminService) CreateScenario(ctx context.Context, input CreateScenarioInput) (*db.Scenario, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, ErrInvalidScenario
	}

	if input.Jurisdiction == "" {
		input.Jurisdiction = "federal"
	}

	if input.PracticeArea == "" {
		input.PracticeArea = "civil"
	}

	if input.HearingType == "" {
		input.HearingType = "trial_direct_examination"
	}

	if input.Difficulty == "" {
		input.Difficulty = "beginner"
	}

	if input.Status == "" {
		input.Status = "draft"
	}

	scenario := &db.Scenario{
		ID:           uuid.NewString(),
		Title:        input.Title,
		Description:  input.Description,
		Jurisdiction: input.Jurisdiction,
		PracticeArea: input.PracticeArea,
		HearingType:  input.HearingType,
		Difficulty:   input.Difficulty,
		Status:       input.Status,
	}

	if err := s.repo.CreateScenario(ctx, scenario); err != nil {
		return nil, err
	}

	return scenario, nil
}

func (s *AdminService) UpdateScenario(
	ctx context.Context,
	scenarioID string,
	input UpdateScenarioInput,
) (*db.Scenario, error) {
	scenario, err := s.repo.GetByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.Title) != "" {
		scenario.Title = input.Title
	}

	scenario.Description = input.Description

	if input.Jurisdiction != "" {
		scenario.Jurisdiction = input.Jurisdiction
	}

	if input.PracticeArea != "" {
		scenario.PracticeArea = input.PracticeArea
	}

	if input.HearingType != "" {
		scenario.HearingType = input.HearingType
	}

	if input.Difficulty != "" {
		scenario.Difficulty = input.Difficulty
	}

	if input.Status != "" {
		scenario.Status = input.Status
	}

	if err := s.repo.UpdateScenario(ctx, scenario); err != nil {
		return nil, err
	}

	return scenario, nil
}

func (s *AdminService) PublishScenario(ctx context.Context, scenarioID string) (*db.Scenario, error) {
	return s.UpdateScenario(ctx, scenarioID, UpdateScenarioInput{
		Status: "published",
	})
}

func (s *AdminService) ArchiveScenario(ctx context.Context, scenarioID string) (*db.Scenario, error) {
	return s.UpdateScenario(ctx, scenarioID, UpdateScenarioInput{
		Status: "archived",
	})
}

func (s *AdminService) CreateScenarioLine(
	ctx context.Context,
	input CreateScenarioLineInput,
) (*db.ScenarioLine, error) {
	if input.ScenarioID == "" || input.SequenceNo <= 0 || strings.TrimSpace(input.LineText) == "" {
		return nil, ErrInvalidScenarioLine
	}

	if input.SpeakerType == "" {
		input.SpeakerType = "witness"
	}

	if input.LineKind == "" {
		input.LineKind = "answer"
	}

	line := &db.ScenarioLine{
		ID:          uuid.NewString(),
		ScenarioID:  input.ScenarioID,
		SequenceNo:  input.SequenceNo,
		SpeakerType: input.SpeakerType,
		SpeakerName: input.SpeakerName,
		LineText:    input.LineText,
		LineKind:    input.LineKind,
	}

	if err := s.repo.CreateScenarioLine(ctx, line); err != nil {
		return nil, err
	}

	return line, nil
}

func (s *AdminService) ListObjectionTypes(ctx context.Context) ([]db.ObjectionType, error) {
	return s.repo.ListObjectionTypes(ctx)
}

func (s *AdminService) CreateOpportunity(
	ctx context.Context,
	input CreateOpportunityInput,
) (*db.ObjectionOpportunity, error) {
	if input.ScenarioLineID == "" || input.ObjectionTypeID == "" || strings.TrimSpace(input.Explanation) == "" {
		return nil, ErrInvalidOpportunity
	}

	if input.Strength == "" {
		input.Strength = "strong"
	}

	if input.TimingWindow == "" {
		input.TimingWindow = "after_answer"
	}

	opportunity := &db.ObjectionOpportunity{
		ID:               uuid.NewString(),
		ScenarioLineID:   input.ScenarioLineID,
		ObjectionTypeID:  input.ObjectionTypeID,
		Strength:         input.Strength,
		TimingWindow:     input.TimingWindow,
		Explanation:      input.Explanation,
		ExpectedPhrase:   input.ExpectedPhrase,
		IsPrimary:        input.IsPrimary,
	}

	if err := s.repo.CreateObjectionOpportunity(ctx, opportunity); err != nil {
		return nil, err
	}

	return opportunity, nil
}