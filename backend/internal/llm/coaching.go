package llm

import "context"

type CoachingInput struct {
	ScenarioID                  string
	LineText                    string
	SpeakerName                 string
	LineKind                    string
	TraineeAction               string
	NormalizedObjectionTypeID   string
	MatchedOpportunityID        string
	Ruling                      string
	Valid                       bool
	Timely                      bool
	LegalAccuracyScore          float64
	PhrasingScore               float64
	StrategyScore               float64
	DeterministicFeedback       string
	ExpectedObjectionExplanation string
	ExpectedPhrase              string
}

type Coach interface {
	EnhanceFeedback(ctx context.Context, input CoachingInput) (string, error)
}

type NoopCoach struct{}

func NewNoopCoach() *NoopCoach {
	return &NoopCoach{}
}

func (c *NoopCoach) EnhanceFeedback(_ context.Context, input CoachingInput) (string, error) {
	return input.DeterministicFeedback, nil
}