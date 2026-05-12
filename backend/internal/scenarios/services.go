package scenarios

import (
	"context"

	"CallItCureIt/backend/internal/db"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ListPublished(ctx context.Context) ([]db.Scenario, error) {
	return s.repo.ListPublished(ctx)
}

func (s *Service) GetScenario(ctx context.Context, id string) (*db.Scenario, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetTranscript(ctx context.Context, scenarioID string) ([]db.ScenarioLine, error) {
	return s.repo.ListLines(ctx, scenarioID)
}

func (s *Service) ListObjectionTypes(ctx context.Context) ([]db.ObjectionType, error) {
	return s.repo.ListObjectionTypes(ctx)
}