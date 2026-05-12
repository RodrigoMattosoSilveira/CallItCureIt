package scenarios

import (
	"context"

	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
)

type Repository interface {
	ListPublished(ctx context.Context) ([]db.Scenario, error)
	GetByID(ctx context.Context, id string) (*db.Scenario, error)
	ListLines(ctx context.Context, scenarioID string) ([]db.ScenarioLine, error)
	GetLineWithOpportunities(ctx context.Context, lineID string) (*db.ScenarioLine, error)
	ListObjectionTypes(ctx context.Context) ([]db.ObjectionType, error)
}

type GormRepository struct {
	database *gorm.DB
}

func NewGormRepository(database *gorm.DB) *GormRepository {
	return &GormRepository{
		database: database,
	}
}

func (r *GormRepository) ListPublished(ctx context.Context) ([]db.Scenario, error) {
	var scenarios []db.Scenario

	err := r.database.WithContext(ctx).
		Where("status = ?", "published").
		Order("difficulty ASC, title ASC").
		Find(&scenarios).Error

	return scenarios, err
}

func (r *GormRepository) GetByID(ctx context.Context, id string) (*db.Scenario, error) {
	var scenario db.Scenario

	err := r.database.WithContext(ctx).
		Preload("Actors").
		Where("id = ?", id).
		First(&scenario).Error

	if err != nil {
		return nil, err
	}

	return &scenario, nil
}

func (r *GormRepository) ListLines(ctx context.Context, scenarioID string) ([]db.ScenarioLine, error) {
	var lines []db.ScenarioLine

	err := r.database.WithContext(ctx).
		Where("scenario_id = ?", scenarioID).
		Order("sequence_no ASC").
		Find(&lines).Error

	return lines, err
}

func (r *GormRepository) GetLineWithOpportunities(ctx context.Context, lineID string) (*db.ScenarioLine, error) {
	var line db.ScenarioLine

	err := r.database.WithContext(ctx).
		Preload("Opportunities").
		Preload("Opportunities.ObjectionType").
		Preload("Opportunities.RuleRefs").
		Where("id = ?", lineID).
		First(&line).Error

	if err != nil {
		return nil, err
	}

	return &line, nil
}

func (r *GormRepository) ListObjectionTypes(ctx context.Context) ([]db.ObjectionType, error) {
	var types []db.ObjectionType

	err := r.database.WithContext(ctx).
		Order("name ASC").
		Find(&types).Error

	return types, err
}