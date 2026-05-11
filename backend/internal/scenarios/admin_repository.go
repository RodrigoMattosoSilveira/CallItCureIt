package scenarios

import (
	"context"

	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
)

type AdminRepository interface {
	ListAll(ctx context.Context) ([]db.Scenario, error)
	CreateScenario(ctx context.Context, scenario *db.Scenario) error
	UpdateScenario(ctx context.Context, scenario *db.Scenario) error
	GetByID(ctx context.Context, id string) (*db.Scenario, error)

	CreateScenarioLine(ctx context.Context, line *db.ScenarioLine) error
	ListLines(ctx context.Context, scenarioID string) ([]db.ScenarioLine, error)

	ListObjectionTypes(ctx context.Context) ([]db.ObjectionType, error)
	CreateObjectionOpportunity(ctx context.Context, opportunity *db.ObjectionOpportunity) error

	UpdateScenarioLine(ctx context.Context, line *db.ScenarioLine) error
	GetScenarioLineByID(ctx context.Context, lineID string) (*db.ScenarioLine, error)
	DeleteScenarioLine(ctx context.Context, lineID string) error

	GetObjectionOpportunityByID(ctx context.Context, opportunityID string) (*db.ObjectionOpportunity, error)
	UpdateObjectionOpportunity(ctx context.Context, opportunity *db.ObjectionOpportunity) error
	DeleteObjectionOpportunity(ctx context.Context, opportunityID string) error
}

type GormAdminRepository struct {
	database *gorm.DB
}

func NewGormAdminRepository(database *gorm.DB) *GormAdminRepository {
	return &GormAdminRepository{
		database: database,
	}
}

func (r *GormAdminRepository) ListAll(ctx context.Context) ([]db.Scenario, error) {
	var scenarios []db.Scenario

	err := r.database.WithContext(ctx).
		Order("created_at DESC").
		Find(&scenarios).Error

	return scenarios, err
}

func (r *GormAdminRepository) CreateScenario(ctx context.Context, scenario *db.Scenario) error {
	return r.database.WithContext(ctx).Create(scenario).Error
}

func (r *GormAdminRepository) UpdateScenario(ctx context.Context, scenario *db.Scenario) error {
	return r.database.WithContext(ctx).Save(scenario).Error
}

func (r *GormAdminRepository) GetByID(ctx context.Context, id string) (*db.Scenario, error) {
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

func (r *GormAdminRepository) CreateScenarioLine(ctx context.Context, line *db.ScenarioLine) error {
	return r.database.WithContext(ctx).Create(line).Error
}

func (r *GormAdminRepository) ListLines(ctx context.Context, scenarioID string) ([]db.ScenarioLine, error) {
	var lines []db.ScenarioLine

	err := r.database.WithContext(ctx).
		Preload("Opportunities").
		Preload("Opportunities.ObjectionType").
		Where("scenario_id = ?", scenarioID).
		Order("sequence_no ASC").
		Find(&lines).Error

	return lines, err
}

func (r *GormAdminRepository) ListObjectionTypes(ctx context.Context) ([]db.ObjectionType, error) {
	var objectionTypes []db.ObjectionType

	err := r.database.WithContext(ctx).
		Order("name ASC").
		Find(&objectionTypes).Error

	return objectionTypes, err
}

func (r *GormAdminRepository) CreateObjectionOpportunity(
	ctx context.Context,
	opportunity *db.ObjectionOpportunity,
) error {
	return r.database.WithContext(ctx).Create(opportunity).Error
}

func (r *GormAdminRepository) GetScenarioLineByID(
	ctx context.Context,
	lineID string,
) (*db.ScenarioLine, error) {
	var line db.ScenarioLine

	err := r.database.WithContext(ctx).
		Where("id = ?", lineID).
		First(&line).Error

	if err != nil {
		return nil, err
	}

	return &line, nil
}

func (r *GormAdminRepository) UpdateScenarioLine(
	ctx context.Context,
	line *db.ScenarioLine,
) error {
	return r.database.WithContext(ctx).Save(line).Error
}

func (r *GormAdminRepository) DeleteScenarioLine(
	ctx context.Context,
	lineID string,
) error {
	return r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`DELETE FROM opportunity_rule_refs
			 WHERE opportunity_id IN (
			   SELECT id FROM objection_opportunities WHERE scenario_line_id = ?
			 )`,
			lineID,
		).Error; err != nil {
			return err
		}

		if err := tx.Exec(
			`DELETE FROM objection_opportunities WHERE scenario_line_id = ?`,
			lineID,
		).Error; err != nil {
			return err
		}

		return tx.Exec(
			`DELETE FROM scenario_lines WHERE id = ?`,
			lineID,
		).Error
	})
}

func (r *GormAdminRepository) GetObjectionOpportunityByID(
	ctx context.Context,
	opportunityID string,
) (*db.ObjectionOpportunity, error) {
	var opportunity db.ObjectionOpportunity

	err := r.database.WithContext(ctx).
		Where("id = ?", opportunityID).
		First(&opportunity).Error

	if err != nil {
		return nil, err
	}

	return &opportunity, nil
}

func (r *GormAdminRepository) UpdateObjectionOpportunity(
	ctx context.Context,
	opportunity *db.ObjectionOpportunity,
) error {
	return r.database.WithContext(ctx).Save(opportunity).Error
}

func (r *GormAdminRepository) DeleteObjectionOpportunity(
	ctx context.Context,
	opportunityID string,
) error {
	return r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`DELETE FROM opportunity_rule_refs WHERE opportunity_id = ?`,
			opportunityID,
		).Error; err != nil {
			return err
		}

		return tx.Exec(
			`DELETE FROM objection_opportunities WHERE id = ?`,
			opportunityID,
		).Error
	})
}