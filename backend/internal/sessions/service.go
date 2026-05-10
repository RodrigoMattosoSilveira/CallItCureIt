package sessions

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/llm"
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
	Session       *db.Session
	Action        *db.TraineeAction
	TraineeEvent  *db.SessionEvent
	JudgeEvent    *db.SessionEvent
	CoachEvent    *db.SessionEvent
	Evaluation    *db.ActionEvaluation
}

type Service struct {
	repo             Repository
	objectionMatcher *objections.Matcher
	coach            llm.Coach
}

type CalculateScoreResult struct {
	Session *db.Session
	Score   *db.SessionScore
}

type DebriefAction struct {
	Action     db.TraineeAction
	Evaluation db.ActionEvaluation
}

type DebriefResult struct {
	Session *db.Session
	Events  []db.SessionEvent
	Actions []DebriefAction
	Score   *db.SessionScore
}

func NewService(repo Repository, coach llm.Coach) *Service {
	if coach == nil {
		coach = llm.NewNoopCoach()
	}

	return &Service{
		repo:             repo,
		objectionMatcher: objections.NewMatcher(),
		coach:            coach,
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
	actionType := strings.TrimSpace(input.ActionType)
	rawText := strings.TrimSpace(input.RawText)

	if actionType == "" {
		actionType = "object"
	}

	if actionType != "object" &&
		actionType != "respond" &&
		actionType != "pass" {
		return nil, ErrInvalidAction
	}

	if actionType != "pass" && rawText == "" {
		return nil, ErrInvalidAction
	}

	if actionType == "pass" && rawText == "" {
		rawText = "Pass"
	}

	session, err := s.repo.GetByID(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}

	if session.Status == "completed" {
		return nil, ErrSessionCompleted
	}

	if session.Status != "active" {
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

	normalized := s.objectionMatcher.Normalize(rawText)

	var normalizedObjectionTypeID *string
	if normalized.Matched {
		id := normalized.ObjectionTypeID
		normalizedObjectionTypeID = &id
	}

	scenarioLineID := line.ID

	action := &db.TraineeAction{
		ID:                        uuid.NewString(),
		SessionID:                 session.ID,
		ScenarioLineID:            &scenarioLineID,
		ActionType:                actionType,
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

	traineeEventType := "trainee_objection"
	if actionType == "respond" {
		traineeEventType = "trainee_response"
	}

	if actionType == "pass" {
		traineeEventType = "trainee_response"
	}

	traineeEvent := &db.SessionEvent{
		ID:         uuid.NewString(),
		SessionID:  session.ID,
		SequenceNo: session.CurrentSequenceNo,
		EventType:  traineeEventType,
		Actor:      "Trainee Counsel",
		Text:       rawText,
	}

	if err := s.repo.CreateEvent(ctx, traineeEvent); err != nil {
		return nil, err
	}

	judgeEvent := &db.SessionEvent{
		ID:         uuid.NewString(),
		SessionID:  session.ID,
		SequenceNo: session.CurrentSequenceNo,
		EventType:  "judge_ruling",
		Actor:      "Judge Carter",
		Text:       buildJudgeRulingText(evaluation),
	}

	if err := s.repo.CreateEvent(ctx, judgeEvent); err != nil {
		return nil, err
	}

	coachEvent := &db.SessionEvent{
		ID:         uuid.NewString(),
		SessionID:  session.ID,
		SequenceNo: session.CurrentSequenceNo,
		EventType:  "coach_feedback",
		Actor:      "Coach",
		Text:       evaluation.Feedback,
	}

	if err := s.repo.CreateEvent(ctx, coachEvent); err != nil {
		return nil, err
	}

	_, _ = s.CalculateScore(ctx, session.ID)

	return &SubmitActionResult{
		Session:      session,
		Action:       action,
		TraineeEvent: traineeEvent,
		JudgeEvent:   judgeEvent,
		CoachEvent:   coachEvent,
		Evaluation:   evaluation,
	}, nil
}

func (s *Service) CalculateScore(
	ctx context.Context,
	sessionID string,
) (*CalculateScoreResult, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	evaluations, err := s.repo.ListActionEvaluations(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	totalOpportunities, err := s.repo.CountOpportunitiesThroughSequence(
		ctx,
		session.ScenarioID,
		session.CurrentSequenceNo,
	)
	if err != nil {
		return nil, err
	}

	matchedOpportunities, err := s.repo.CountMatchedOpportunities(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	falsePositives, err := s.repo.CountFalsePositives(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	evaluatedActionCount := len(evaluations)

	var legalAccuracy float64
	var timeliness float64
	var phrasing float64
	var strategy float64

	if evaluatedActionCount > 0 {
		var legalTotal float64
		var timelyTotal float64
		var phrasingTotal float64
		var strategyTotal float64

		for _, evaluation := range evaluations {
			legalTotal += evaluation.LegalAccuracyScore
			phrasingTotal += evaluation.PhrasingScore
			strategyTotal += evaluation.StrategyScore

			if evaluation.Timely {
				timelyTotal += 100
			}
		}

		legalAccuracy = legalTotal / float64(evaluatedActionCount)
		timeliness = timelyTotal / float64(evaluatedActionCount)
		phrasing = phrasingTotal / float64(evaluatedActionCount)
		strategy = strategyTotal / float64(evaluatedActionCount)
	}

	var spottingAccuracy float64
	if totalOpportunities > 0 {
		spottingAccuracy = float64(matchedOpportunities) / float64(totalOpportunities) * 100
	} else {
		// If no opportunities were encountered yet, do not award 100 by default.
		// This avoids misleading early perfect scores.
		spottingAccuracy = 0
	}

	missedOpportunities := totalOpportunities - matchedOpportunities
	if missedOpportunities < 0 {
		missedOpportunities = 0
	}

	responseQuality := averageNonZero([]float64{
		legalAccuracy,
		phrasing,
		strategy,
	})

	overallScore := weightedOverallScore(
		spottingAccuracy,
		legalAccuracy,
		timeliness,
		phrasing,
		strategy,
	)

	now := time.Now()

	score := &db.SessionScore{
		ID:                      uuid.NewString(),
		SessionID:               session.ID,
		EvaluatedActionCount:    evaluatedActionCount,
		TotalOpportunityCount:   int(totalOpportunities),
		MatchedOpportunityCount: int(matchedOpportunities),
		MissedOpportunityCount:  int(missedOpportunities),
		FalsePositiveCount:      int(falsePositives),

		SpottingAccuracy: roundScore(spottingAccuracy),
		LegalAccuracy:    roundScore(legalAccuracy),
		Timeliness:        roundScore(timeliness),
		Phrasing:          roundScore(phrasing),
		Strategy:          roundScore(strategy),
		ResponseQuality:   roundScore(responseQuality),
		OverallScore:      roundScore(overallScore),

		IsFinal:   session.Status == "completed",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.UpsertSessionScore(ctx, score); err != nil {
		return nil, err
	}

	return &CalculateScoreResult{
		Session: session,
		Score:   score,
	}, nil
}

func (s *Service) GetScore(
	ctx context.Context,
	sessionID string,
) (*CalculateScoreResult, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	score, err := s.repo.GetSessionScore(ctx, sessionID)
	if err != nil {
		if IsNotFound(err) {
			evaluations, evalErr := s.repo.ListActionEvaluationsBySession(ctx, sessionID)
			if evalErr != nil {
				return nil, evalErr
			}

			score = calculateSessionScore(session.ID, evaluations)

			if upsertErr := s.repo.UpsertSessionScore(ctx, score); upsertErr != nil {
				return nil, upsertErr
			}
		} else {
			return nil, err
		}
	}

	return &CalculateScoreResult{
		Session: session,
		Score:   score,
	}, nil
}

func calculateSessionScore(
	sessionID string,
	evaluations []db.ActionEvaluation,
) *db.SessionScore {
	count := len(evaluations)

	score := &db.SessionScore{
		ID:                   uuid.NewString(),
		SessionID:            sessionID,
		EvaluatedActionCount: count,
	}

	if count == 0 {
		score.SpottingAccuracy = 0
		score.LegalAccuracy = 0
		score.Timeliness = 0
		score.Phrasing = 0
		score.ResponseQuality = 0
		score.Strategy = 0
		score.OverallScore = 0
		return score
	}

	var legalTotal float64
	var phrasingTotal float64
	var strategyTotal float64
	var timelyCount float64
	var validCount float64

	for _, evaluation := range evaluations {
		legalTotal += evaluation.LegalAccuracyScore
		phrasingTotal += evaluation.PhrasingScore
		strategyTotal += evaluation.StrategyScore

		if evaluation.Timely {
			timelyCount++
		}

		if evaluation.Valid {
			validCount++
		}
	}

	score.LegalAccuracy = roundScore(legalTotal / float64(count))
	score.Phrasing = roundScore(phrasingTotal / float64(count))
	score.Strategy = roundScore(strategyTotal / float64(count))
	score.Timeliness = roundScore((timelyCount / float64(count)) * 100)
	score.SpottingAccuracy = roundScore((validCount / float64(count)) * 100)

	// Response quality is mostly relevant for the future respond_to_objection mode.
	// For now, use legal accuracy as a reasonable MVP proxy.
	score.ResponseQuality = score.LegalAccuracy

	score.OverallScore = roundScore(
		(score.SpottingAccuracy * 0.25) +
			(score.LegalAccuracy * 0.30) +
			(score.Timeliness * 0.15) +
			(score.Phrasing * 0.10) +
			(score.Strategy * 0.20),
	)

	return score
}

func roundScore(value float64) float64 {
	return math.Round(value*100) / 100
}

// Full deterministic evaluation helper
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
		return evaluatePassAction(evaluation, line)
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

		evaluation.Valid = false
		evaluation.Timely = true
		evaluation.Ruling = "overruled"
		evaluation.LegalAccuracyScore = 25
		evaluation.StrategyScore = 25
		evaluation.Feedback = "This line had an objection opportunity, but your objection ground did not match the strongest expected ground. Expected: " +
			expected.ObjectionType.Name + "."

		return evaluation
	}

	evaluation.Valid = false
	evaluation.Timely = false
	evaluation.Ruling = "overruled"
	evaluation.LegalAccuracyScore = 0
	evaluation.StrategyScore = 0
	evaluation.Feedback = "There was no expected objection opportunity on this line, so the objection would likely be overruled."

	return evaluation
}

// Pass-action helper
func evaluatePassAction(
	evaluation *db.ActionEvaluation,
	line *db.ScenarioLine,
) *db.ActionEvaluation {
	evaluation.Ruling = "no_ruling"
	evaluation.PhrasingScore = 100

	if len(line.Opportunities) > 0 {
		expected := line.Opportunities[0]

		evaluation.Valid = false
		evaluation.Timely = false
		evaluation.LegalAccuracyScore = 0
		evaluation.StrategyScore = 0
		evaluation.Feedback = "You passed, but this line had an objection opportunity. Expected: " +
			expected.ObjectionType.Name + "."

		return evaluation
	}

	evaluation.Valid = true
	evaluation.Timely = true
	evaluation.LegalAccuracyScore = 100
	evaluation.StrategyScore = 100
	evaluation.Feedback = "Correct. There was no strong objection opportunity on this line."

	return evaluation
}

// Scoring helpers
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

// Feedback and judge helpers
func buildCorrectFeedback(opportunity db.ObjectionOpportunity) string {
	if opportunity.Explanation != "" {
		return "Correct. " + opportunity.Explanation
	}

	if opportunity.ObjectionType.Name != "" {
		return "Correct. The objection matches the expected ground: " +
			opportunity.ObjectionType.Name + "."
	}

	return "Correct. The objection matches the expected opportunity."
}

func buildJudgeRulingText(evaluation *db.ActionEvaluation) string {
	switch evaluation.Ruling {
	case "sustained":
		return "Sustained."
	case "overruled":
		return "Overruled."
	case "no_ruling":
		return "No ruling."
	default:
		return "The court will note the objection."
	}
}

func weightedOverallScore(
	spottingAccuracy float64,
	legalAccuracy float64,
	timeliness float64,
	phrasing float64,
	strategy float64,
) float64 {
	return (spottingAccuracy * 0.40) +
		(legalAccuracy * 0.25) +
		(timeliness * 0.15) +
		(phrasing * 0.10) +
		(strategy * 0.10)
}

func averageNonZero(values []float64) float64 {
	var total float64
	var count float64

	for _, value := range values {
		if value > 0 {
			total += value
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / count
}

func (s *Service) GetDebrief(
	ctx context.Context,
	sessionID string,
) (*DebriefResult, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.ListEvents(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	actions, err := s.repo.ListTraineeActionsWithEvaluations(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	scoreResult, err := s.GetScore(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	debriefActions := make([]DebriefAction, 0, len(actions))

	for _, action := range actions {
		if action.Evaluation == nil {
			continue
		}

		debriefActions = append(debriefActions, DebriefAction{
			Action:     action,
			Evaluation: *action.Evaluation,
		})
	}

	return &DebriefResult{
		Session: session,
		Events:  events,
		Actions: debriefActions,
		Score:   scoreResult.Score,
	}, nil
}

func (s *Service) buildEnhancedCoachFeedback(
	ctx context.Context,
	session *db.Session,
	line *db.ScenarioLine,
	action *db.TraineeAction,
	evaluation *db.ActionEvaluation,
) string {
	input := llm.CoachingInput{
		ScenarioID:            session.ScenarioID,
		LineText:              line.LineText,
		SpeakerName:           line.SpeakerName,
		LineKind:              line.LineKind,
		TraineeAction:         action.RawText,
		Ruling:                evaluation.Ruling,
		Valid:                 evaluation.Valid,
		Timely:                evaluation.Timely,
		LegalAccuracyScore:    evaluation.LegalAccuracyScore,
		PhrasingScore:         evaluation.PhrasingScore,
		StrategyScore:         evaluation.StrategyScore,
		DeterministicFeedback: evaluation.Feedback,
	}

	if evaluation.NormalizedObjectionTypeID != nil {
		input.NormalizedObjectionTypeID = *evaluation.NormalizedObjectionTypeID
	}

	if evaluation.MatchedOpportunityID != nil {
		input.MatchedOpportunityID = *evaluation.MatchedOpportunityID
	}

	for _, opportunity := range line.Opportunities {
		if evaluation.MatchedOpportunityID != nil && opportunity.ID == *evaluation.MatchedOpportunityID {
			input.ExpectedObjectionExplanation = opportunity.Explanation
			input.ExpectedPhrase = opportunity.ExpectedPhrase
			break
		}
	}

	if input.ExpectedObjectionExplanation == "" && len(line.Opportunities) > 0 {
		input.ExpectedObjectionExplanation = line.Opportunities[0].Explanation
		input.ExpectedPhrase = line.Opportunities[0].ExpectedPhrase
	}

	enhanced, err := s.coach.EnhanceFeedback(ctx, input)
	if err != nil {
		return evaluation.Feedback
	}

	enhanced = strings.TrimSpace(enhanced)
	if enhanced == "" {
		return evaluation.Feedback
	}

	return enhanced
}