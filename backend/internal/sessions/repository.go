package sessions

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
)

type Repository interface {
	Create(ctx context.Context, session *db.Session) error
	GetByID(ctx context.Context, id string) (*db.Session, error)
	Update(ctx context.Context, session *db.Session) error

	GetNextScenarioLine(ctx context.Context, scenarioID string, afterSequenceNo int) (*db.ScenarioLine, error)
	GetScenarioLineBySequence(ctx context.Context, scenarioID string, sequenceNo int) (*db.ScenarioLine, error)
	GetScenarioLineWithOpportunities(ctx context.Context, scenarioID string, sequenceNo int) (*db.ScenarioLine, error)

	CreateEvent(ctx context.Context, event *db.SessionEvent) error
	ListEvents(ctx context.Context, sessionID string) ([]db.SessionEvent, error)
	CountScenarioLines(ctx context.Context, scenarioID string) (int64, error)

	CreateTraineeAction(ctx context.Context, action *db.TraineeAction) error
	ListTraineeActions(ctx context.Context, sessionID string) ([]db.TraineeAction, error)

	CreateActionEvaluation(ctx context.Context, evaluation *db.ActionEvaluation) error
	ListActionEvaluationsBySession(ctx context.Context, sessionID string) ([]db.ActionEvaluation, error)

	ListActionEvaluations(ctx context.Context, sessionID string) ([]db.ActionEvaluation, error)
	UpsertSessionScore(ctx context.Context, score *db.SessionScore) error
	GetSessionScore(ctx context.Context, sessionID string) (*db.SessionScore, error)

	CountOpportunitiesThroughSequence(ctx context.Context, scenarioID string, sequenceNo int) (int64, error)
	CountMatchedOpportunities(ctx context.Context, sessionID string) (int64, error)
	CountFalsePositives(ctx context.Context, sessionID string) (int64, error)
}

type GormRepository struct {
	database *gorm.DB
}

func NewGormRepository(database *gorm.DB) *GormRepository {
	return &GormRepository{
		database: database,
	}
}

func (r *GormRepository) Create(ctx context.Context, session *db.Session) error {
	return r.database.WithContext(ctx).Create(session).Error
}

func (r *GormRepository) GetByID(ctx context.Context, id string) (*db.Session, error) {
	var session db.Session

	err := r.database.WithContext(ctx).
		Preload("Scenario").
		Where("id = ?", id).
		First(&session).Error

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *GormRepository) Update(ctx context.Context, session *db.Session) error {
	return r.database.WithContext(ctx).Save(session).Error
}

func (r *GormRepository) GetNextScenarioLine(
	ctx context.Context,
	scenarioID string,
	afterSequenceNo int,
) (*db.ScenarioLine, error) {
	var line db.ScenarioLine

	err := r.database.WithContext(ctx).
		Where("scenario_id = ? AND sequence_no > ?", scenarioID, afterSequenceNo).
		Order("sequence_no ASC").
		First(&line).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoMoreLines
	}

	if err != nil {
		return nil, err
	}

	return &line, nil
}

func (r *GormRepository) CreateEvent(ctx context.Context, event *db.SessionEvent) error {
	return r.database.WithContext(ctx).Create(event).Error
}

func (r *GormRepository) ListEvents(ctx context.Context, sessionID string) ([]db.SessionEvent, error) {
	var events []db.SessionEvent

	err := r.database.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("sequence_no ASC, created_at ASC").
		Find(&events).Error

	return events, err
}

func (r *GormRepository) CountScenarioLines(ctx context.Context, scenarioID string) (int64, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Model(&db.ScenarioLine{}).
		Where("scenario_id = ?", scenarioID).
		Count(&count).Error

	return count, err
}

func (r *GormRepository) GetScenarioLineBySequence(
	ctx context.Context,
	scenarioID string,
	sequenceNo int,
) (*db.ScenarioLine, error) {
	var line db.ScenarioLine

	err := r.database.WithContext(ctx).
		Where("scenario_id = ? AND sequence_no = ?", scenarioID, sequenceNo).
		First(&line).Error

	if err != nil {
		return nil, err
	}

	return &line, nil
}

func (r *GormRepository) CreateTraineeAction(ctx context.Context, action *db.TraineeAction) error {
	return r.database.WithContext(ctx).Create(action).Error
}

func (r *GormRepository) ListTraineeActions(
	ctx context.Context,
	sessionID string,
) ([]db.TraineeAction, error) {
	var actions []db.TraineeAction

	err := r.database.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&actions).Error

	return actions, err
}

func (r *GormRepository) GetScenarioLineWithOpportunities(
	ctx context.Context,
	scenarioID string,
	sequenceNo int,
) (*db.ScenarioLine, error) {
	var line db.ScenarioLine

	err := r.database.WithContext(ctx).
		Preload("Opportunities").
		Preload("Opportunities.ObjectionType").
		Preload("Opportunities.RuleRefs").
		Where("scenario_id = ? AND sequence_no = ?", scenarioID, sequenceNo).
		First(&line).Error

	if err != nil {
		return nil, err
	}

	return &line, nil
}

func (r *GormRepository) CreateActionEvaluation(
	ctx context.Context,
	evaluation *db.ActionEvaluation,
) error {
	return r.database.WithContext(ctx).Create(evaluation).Error
}

func (r *GormRepository) ListActionEvaluationsBySession(
	ctx context.Context,
	sessionID string,
) ([]db.ActionEvaluation, error) {
	var evaluations []db.ActionEvaluation

	err := r.database.WithContext(ctx).
		Joins("JOIN trainee_actions ON trainee_actions.id = action_evaluations.trainee_action_id").
		Where("trainee_actions.session_id = ?", sessionID).
		Order("action_evaluations.created_at ASC").
		Find(&evaluations).Error

	return evaluations, err
}

func (r *GormRepository) UpsertSessionScore(
	ctx context.Context,
	score *db.SessionScore,
) error {
	var existing db.SessionScore

	err := r.database.WithContext(ctx).
		Where("session_id = ?", score.SessionID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.database.WithContext(ctx).Create(score).Error
	}

	if err != nil {
		return err
	}

	score.ID = existing.ID
	score.CreatedAt = existing.CreatedAt

	return r.database.WithContext(ctx).Save(score).Error
}

func (r *GormRepository) GetSessionScore(
	ctx context.Context,
	sessionID string,
) (*db.SessionScore, error) {
	var score db.SessionScore

	err := r.database.WithContext(ctx).
		Where("session_id = ?", sessionID).
		First(&score).Error

	if err != nil {
		return nil, err
	}

	return &score, nil
}

func (r *GormRepository) CountOpportunitiesThroughSequence(
	ctx context.Context,
	scenarioID string,
	sequenceNo int,
) (int64, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Table("objection_opportunities oo").
		Joins("JOIN scenario_lines sl ON sl.id = oo.scenario_line_id").
		Where("sl.scenario_id = ? AND sl.sequence_no <= ?", scenarioID, sequenceNo).
		Count(&count).Error

	return count, err
}

func (r *GormRepository) CountMatchedOpportunities(
	ctx context.Context,
	sessionID string,
) (int64, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Table("action_evaluations ae").
		Joins("JOIN trainee_actions ta ON ta.id = ae.trainee_action_id").
		Where("ta.session_id = ? AND ae.valid = 1 AND ae.matched_opportunity_id IS NOT NULL", sessionID).
		Distinct("ae.matched_opportunity_id").
		Count(&count).Error

	return count, err
}

func (r *GormRepository) CountFalsePositives(
	ctx context.Context,
	sessionID string,
) (int64, error) {
	var count int64

	err := r.database.WithContext(ctx).
		Table("action_evaluations ae").
		Joins("JOIN trainee_actions ta ON ta.id = ae.trainee_action_id").
		Where("ta.session_id = ? AND ae.valid = 0 AND ta.action_type = ?", sessionID, "object").
		Count(&count).Error

	return count, err
}

func (r *GormRepository) ListActionEvaluations(
	ctx context.Context,
	sessionID string,
) ([]db.ActionEvaluation, error) {
	var evaluations []db.ActionEvaluation

	err := r.database.WithContext(ctx).
		Table("action_evaluations").
		Joins("JOIN trainee_actions ON trainee_actions.id = action_evaluations.trainee_action_id").
		Where("trainee_actions.session_id = ?", sessionID).
		Order("action_evaluations.created_at ASC").
		Find(&evaluations).Error

	return evaluations, err
}