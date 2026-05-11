package sessions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	appdb "CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/llm"
)

func TestHearsayObjectionTrainingFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	database := newTestDatabase(t)

	repo := NewGormRepository(database)
	service := NewService(repo, llm.NewNoopCoach())

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ScenarioID: "scenario-hearsay-001",
		Mode:       "spot_objection",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if session.ID == "" {
		t.Fatal("expected session ID to be set")
	}

	var advanceResult *AdvanceSessionResult

	for i := 0; i < 4; i++ {
		advanceResult, err = service.AdvanceSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("advance session step %d: %v", i+1, err)
		}

		if advanceResult.Completed {
			t.Fatalf("session completed too early at step %d", i+1)
		}

		if advanceResult.Line == nil {
			t.Fatalf("expected line at step %d", i+1)
		}
	}

	if advanceResult.Line.ID != "line-hearsay-004" {
		t.Fatalf(
			"expected to advance to hearsay line line-hearsay-004, got %q",
			advanceResult.Line.ID,
		)
	}

	if !strings.Contains(strings.ToLower(advanceResult.Line.LineText), "defendant admitted") {
		t.Fatalf("expected hearsay line text, got %q", advanceResult.Line.LineText)
	}

	submitResult, err := service.SubmitAction(ctx, SubmitActionInput{
		SessionID:  session.ID,
		ActionType: "object",
		RawText:    "Objection, hearsay.",
	})
	if err != nil {
		t.Fatalf("submit objection: %v", err)
	}

	if submitResult.Evaluation == nil {
		t.Fatal("expected evaluation")
	}

	if !submitResult.Evaluation.Valid {
		t.Fatalf("expected evaluation.valid = true, got false; feedback=%q", submitResult.Evaluation.Feedback)
	}

	if submitResult.Evaluation.Ruling != "sustained" {
		t.Fatalf("expected ruling sustained, got %q", submitResult.Evaluation.Ruling)
	}

	if submitResult.JudgeEvent == nil {
		t.Fatal("expected judge event")
	}

	if submitResult.JudgeEvent.EventType != "judge_ruling" {
		t.Fatalf("expected judge_ruling event, got %q", submitResult.JudgeEvent.EventType)
	}

	if submitResult.JudgeEvent.Text != "Sustained." {
		t.Fatalf("expected judge event text Sustained., got %q", submitResult.JudgeEvent.Text)
	}

	if submitResult.CoachEvent == nil {
		t.Fatal("expected coach event")
	}

	if submitResult.CoachEvent.EventType != "coach_feedback" {
		t.Fatalf("expected coach_feedback event, got %q", submitResult.CoachEvent.EventType)
	}

	if strings.TrimSpace(submitResult.CoachEvent.Text) == "" {
		t.Fatal("expected coach event text")
	}

	scoreResult, err := service.GetScore(ctx, session.ID)
	if err != nil {
		t.Fatalf("get score: %v", err)
	}

	if scoreResult.Score == nil {
		t.Fatal("expected score")
	}

	if scoreResult.Score.OverallScore <= 0 {
		t.Fatalf("expected overall score > 0, got %.2f", scoreResult.Score.OverallScore)
	}

	if scoreResult.Score.EvaluatedActionCount != 1 {
		t.Fatalf(
			"expected evaluated action count 1, got %d",
			scoreResult.Score.EvaluatedActionCount,
		)
	}

	debrief, err := service.GetDebrief(ctx, session.ID)
	if err != nil {
		t.Fatalf("get debrief: %v", err)
	}

	if debrief == nil {
		t.Fatal("expected debrief")
	}

	if len(debrief.Actions) == 0 {
		t.Fatal("expected debrief to include at least one action")
	}

	foundSubmittedAction := false

	for _, item := range debrief.Actions {
		if item.Action.RawText == "Objection, hearsay." {
			foundSubmittedAction = true

			if !item.Evaluation.Valid {
				t.Fatal("expected debrief action evaluation to be valid")
			}

			if item.Evaluation.Ruling != "sustained" {
				t.Fatalf(
					"expected debrief action ruling sustained, got %q",
					item.Evaluation.Ruling,
				)
			}
		}
	}

	if !foundSubmittedAction {
		t.Fatal("expected debrief to include submitted hearsay objection")
	}
}

func newTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "app.db")

	database, err := appdb.ConnectSQLite(dbPath)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	applyMigrations(t, database)

	return database
}

func applyMigrations(t *testing.T, database *gorm.DB) {
	t.Helper()

	migrations := []string{
		"000001_init_schema.up.sql",
		"000002_seed_reference_data.up.sql",
		"000003_create_sessions.up.sql",
		"000004_create_trainee_actions.up.sql",
		"000005_create_action_evaluations.up.sql",
		"000006_create_session_scores.up.sql",
	}

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	for _, migration := range migrations {
		path := filepath.Join("..", "..", "migrations", migration)

		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", migration, err)
		}

		if _, err := sqlDB.Exec(string(contents)); err != nil {
			t.Fatalf("apply migration %s: %v", migration, err)
		}
	}
}

func TestSubmitCorrectHearsayObjectionOnHearsayLine(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	database := newTestDatabase(t)

	repo := NewGormRepository(database)
	service := NewService(repo, llm.NewNoopCoach())

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ScenarioID: "scenario-hearsay-001",
		Mode:       "spot_objection",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Advance to line-hearsay-004, the line with the hearsay opportunity.
	var advanceResult *AdvanceSessionResult

	for i := range 4 {
		advanceResult, err = service.AdvanceSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("advance session step %d: %v", i+1, err)
		}
	}

	if advanceResult.Line == nil {
		t.Fatal("expected current line")
	}

	if advanceResult.Line.ID != "line-hearsay-004" {
		t.Fatalf("expected line-hearsay-004, got %q", advanceResult.Line.ID)
	}

	result, err := service.SubmitAction(ctx, SubmitActionInput{
		SessionID:  session.ID,
		ActionType: "object",
		RawText:    "Objection, hearsay.",
	})
	if err != nil {
		t.Fatalf("submit correct hearsay objection: %v", err)
	}

	if result.Action == nil {
		t.Fatal("expected trainee action")
	}

	if result.Action.RawText != "Objection, hearsay." {
		t.Fatalf("expected raw objection text, got %q", result.Action.RawText)
	}

	if result.Action.NormalizedObjectionTypeID == nil {
		t.Fatal("expected normalized objection type id")
	}

	if *result.Action.NormalizedObjectionTypeID != "obj-hearsay" {
		t.Fatalf(
			"expected normalized objection type obj-hearsay, got %q",
			*result.Action.NormalizedObjectionTypeID,
		)
	}

	if result.Evaluation == nil {
		t.Fatal("expected action evaluation")
	}

	if !result.Evaluation.Valid {
		t.Fatalf("expected valid evaluation; feedback=%q", result.Evaluation.Feedback)
	}

	if !result.Evaluation.Timely {
		t.Fatal("expected timely evaluation")
	}

	if result.Evaluation.Ruling != "sustained" {
		t.Fatalf("expected ruling sustained, got %q", result.Evaluation.Ruling)
	}

	if result.Evaluation.MatchedOpportunityID == nil {
		t.Fatal("expected matched opportunity id")
	}

	if *result.Evaluation.MatchedOpportunityID != "opp-hearsay-001" {
		t.Fatalf(
			"expected matched opportunity opp-hearsay-001, got %q",
			*result.Evaluation.MatchedOpportunityID,
		)
	}

	if result.Evaluation.NormalizedObjectionTypeID == nil {
		t.Fatal("expected evaluation normalized objection type id")
	}

	if *result.Evaluation.NormalizedObjectionTypeID != "obj-hearsay" {
		t.Fatalf(
			"expected evaluation normalized objection type obj-hearsay, got %q",
			*result.Evaluation.NormalizedObjectionTypeID,
		)
	}

	if result.Evaluation.LegalAccuracyScore <= 0 {
		t.Fatalf(
			"expected legal accuracy score > 0, got %.2f",
			result.Evaluation.LegalAccuracyScore,
		)
	}

	if result.JudgeEvent == nil {
		t.Fatal("expected judge event")
	}

	if result.JudgeEvent.EventType != "judge_ruling" {
		t.Fatalf("expected judge_ruling event, got %q", result.JudgeEvent.EventType)
	}

	if result.JudgeEvent.Text != "Sustained." {
		t.Fatalf("expected judge text Sustained., got %q", result.JudgeEvent.Text)
	}

	if result.CoachEvent == nil {
		t.Fatal("expected coach event")
	}

	if result.CoachEvent.EventType != "coach_feedback" {
		t.Fatalf("expected coach_feedback event, got %q", result.CoachEvent.EventType)
	}

	if strings.TrimSpace(result.CoachEvent.Text) == "" {
		t.Fatal("expected non-empty coach feedback")
	}
}

/**
 * This verifies: wrong objection on the hearsay line. The objection is still 
 * expected to be recognized as relevance, but the evaluation should indicate 
 * that it's not valid, not timely, and the ruling should be overruled. The 
 * feedback should indicate that the expected objection was hearsay.
 */
func TestSubmitWrongObjectionOnHearsayLine(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	database := newTestDatabase(t)

	repo := NewGormRepository(database)
	service := NewService(repo, llm.NewNoopCoach())

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ScenarioID: "scenario-hearsay-001",
		Mode:       "spot_objection",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Advance to line-hearsay-004, the line with the hearsay opportunity.
	var advanceResult *AdvanceSessionResult

	for i := range 4 {
		advanceResult, err = service.AdvanceSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("advance session step %d: %v", i+1, err)
		}
	}

	if advanceResult.Line == nil {
		t.Fatal("expected current line")
	}

	if advanceResult.Line.ID != "line-hearsay-004" {
		t.Fatalf("expected line-hearsay-004, got %q", advanceResult.Line.ID)
	}

	result, err := service.SubmitAction(ctx, SubmitActionInput{
		SessionID:  session.ID,
		ActionType: "object",
		RawText:    "Objection, relevance.",
	})
	if err != nil {
		t.Fatalf("submit wrong objection: %v", err)
	}

	if result.Action == nil {
		t.Fatal("expected trainee action")
	}

	if result.Action.RawText != "Objection, relevance." {
		t.Fatalf("expected raw objection text, got %q", result.Action.RawText)
	}

	if result.Action.NormalizedObjectionTypeID == nil {
		t.Fatal("expected normalized objection type id")
	}

	if *result.Action.NormalizedObjectionTypeID != "obj-relevance" {
		t.Fatalf(
			"expected normalized objection type obj-relevance, got %q",
			*result.Action.NormalizedObjectionTypeID,
		)
	}

	if result.Evaluation == nil {
		t.Fatal("expected action evaluation")
	}

	if result.Evaluation.Valid {
		t.Fatalf("expected invalid evaluation; feedback=%q", result.Evaluation.Feedback)
	}

	if !result.Evaluation.Timely {
		t.Fatal("expected timely=true because the objection was made on the correct line")
	}

	if result.Evaluation.Ruling != "overruled" {
		t.Fatalf("expected ruling overruled, got %q", result.Evaluation.Ruling)
	}

	if result.Evaluation.MatchedOpportunityID != nil {
		t.Fatalf(
			"expected no matched opportunity id, got %q",
			*result.Evaluation.MatchedOpportunityID,
		)
	}

	if result.Evaluation.NormalizedObjectionTypeID == nil {
		t.Fatal("expected evaluation normalized objection type id")
	}

	if *result.Evaluation.NormalizedObjectionTypeID != "obj-relevance" {
		t.Fatalf(
			"expected evaluation normalized objection type obj-relevance, got %q",
			*result.Evaluation.NormalizedObjectionTypeID,
		)
	}

	if result.Evaluation.LegalAccuracyScore >= 100 {
		t.Fatalf(
			"expected legal accuracy score below 100 for wrong objection, got %.2f",
			result.Evaluation.LegalAccuracyScore,
		)
	}

	if result.Evaluation.StrategyScore >= 100 {
		t.Fatalf(
			"expected strategy score below 100 for wrong objection, got %.2f",
			result.Evaluation.StrategyScore,
		)
	}

	if !strings.Contains(result.Evaluation.Feedback, "Expected: Hearsay") {
		t.Fatalf(
			"expected feedback to mention Expected: Hearsay, got %q",
			result.Evaluation.Feedback,
		)
	}

	if result.JudgeEvent == nil {
		t.Fatal("expected judge event")
	}

	if result.JudgeEvent.EventType != "judge_ruling" {
		t.Fatalf("expected judge_ruling event, got %q", result.JudgeEvent.EventType)
	}

	if result.JudgeEvent.Text != "Overruled." {
		t.Fatalf("expected judge text Overruled., got %q", result.JudgeEvent.Text)
	}

	if result.CoachEvent == nil {
		t.Fatal("expected coach event")
	}

	if result.CoachEvent.EventType != "coach_feedback" {
		t.Fatalf("expected coach_feedback event, got %q", result.CoachEvent.EventType)
	}

	if strings.TrimSpace(result.CoachEvent.Text) == "" {
		t.Fatal("expected non-empty coach feedback")
	}

	if !strings.Contains(result.CoachEvent.Text, "Expected: Hearsay") {
		t.Fatalf(
			"expected coach feedback to mention Expected: Hearsay, got %q",
			result.CoachEvent.Text,
		)
	}
}

func TestPassOnHearsayLine(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	database := newTestDatabase(t)

	repo := NewGormRepository(database)
	service := NewService(repo, llm.NewNoopCoach())

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ScenarioID: "scenario-hearsay-001",
		Mode:       "spot_objection",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Advance to line-hearsay-004, the line with the hearsay opportunity.
	var advanceResult *AdvanceSessionResult

	for i := range 4 {
		advanceResult, err = service.AdvanceSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("advance session step %d: %v", i+1, err)
		}
	}

	if advanceResult.Line == nil {
		t.Fatal("expected current line")
	}

	if advanceResult.Line.ID != "line-hearsay-004" {
		t.Fatalf("expected line-hearsay-004, got %q", advanceResult.Line.ID)
	}

	result, err := service.SubmitAction(ctx, SubmitActionInput{
		SessionID:  session.ID,
		ActionType: "pass",
		RawText:    "Pass",
	})
	if err != nil {
		t.Fatalf("submit pass on hearsay line: %v", err)
	}

	if result.Action == nil {
		t.Fatal("expected trainee action")
	}

	if result.Action.ActionType != "pass" {
		t.Fatalf("expected action type pass, got %q", result.Action.ActionType)
	}

	if result.Action.RawText != "Pass" {
		t.Fatalf("expected raw text Pass, got %q", result.Action.RawText)
	}

	if result.Action.NormalizedObjectionTypeID != nil {
		t.Fatalf(
			"expected no normalized objection type for pass, got %q",
			*result.Action.NormalizedObjectionTypeID,
		)
	}

	if result.Evaluation == nil {
		t.Fatal("expected action evaluation")
	}

	if result.Evaluation.Valid {
		t.Fatalf("expected pass on hearsay line to be invalid; feedback=%q", result.Evaluation.Feedback)
	}

	if result.Evaluation.Timely {
		t.Fatal("expected timely=false because trainee passed on an objection opportunity")
	}

	if result.Evaluation.Ruling != "no_ruling" {
		t.Fatalf("expected ruling no_ruling, got %q", result.Evaluation.Ruling)
	}

	if result.Evaluation.MatchedOpportunityID != nil {
		t.Fatalf(
			"expected no matched opportunity id for pass, got %q",
			*result.Evaluation.MatchedOpportunityID,
		)
	}

	if result.Evaluation.NormalizedObjectionTypeID != nil {
		t.Fatalf(
			"expected no evaluation normalized objection type for pass, got %q",
			*result.Evaluation.NormalizedObjectionTypeID,
		)
	}

	if result.Evaluation.LegalAccuracyScore != 0 {
		t.Fatalf(
			"expected legal accuracy score 0 for missed opportunity, got %.2f",
			result.Evaluation.LegalAccuracyScore,
		)
	}

	if result.Evaluation.StrategyScore != 0 {
		t.Fatalf(
			"expected strategy score 0 for missed opportunity, got %.2f",
			result.Evaluation.StrategyScore,
		)
	}

	if !strings.Contains(result.Evaluation.Feedback, "You passed") {
		t.Fatalf(
			"expected feedback to mention passed, got %q",
			result.Evaluation.Feedback,
		)
	}

	if result.JudgeEvent == nil {
		t.Fatal("expected judge event")
	}

	if result.JudgeEvent.EventType != "judge_ruling" {
		t.Fatalf("expected judge_ruling event, got %q", result.JudgeEvent.EventType)
	}

	if result.JudgeEvent.Text != "No ruling." {
		t.Fatalf("expected judge text No ruling., got %q", result.JudgeEvent.Text)
	}

	if result.CoachEvent == nil {
		t.Fatal("expected coach event")
	}

	if result.CoachEvent.EventType != "coach_feedback" {
		t.Fatalf("expected coach_feedback event, got %q", result.CoachEvent.EventType)
	}

	if strings.TrimSpace(result.CoachEvent.Text) == "" {
		t.Fatal("expected non-empty coach feedback")
	}

	if !strings.Contains(result.CoachEvent.Text, "You passed") {
		t.Fatalf(
			"expected coach feedback to mention passed, got %q",
			result.CoachEvent.Text,
		)
	}
}

func TestFalsePositiveObjectionOnCleanLine(t *testing.T) {
	t.Parallel()

	h := newSessionTestHarness(t)

	// Advance to line-hearsay-001, a clean line with no objection opportunity.
	advanceToLine(t, h, 1, "line-hearsay-001")

	result := submitTraineeAction(t, h, "object", "Objection, hearsay.")

	if result.Action.RawText != "Objection, hearsay." {
		t.Fatalf("expected raw objection text, got %q", result.Action.RawText)
	}

	requireNormalizedObjectionType(
		t,
		result.Action.NormalizedObjectionTypeID,
		"obj-hearsay",
	)

	if result.Evaluation.Valid {
		t.Fatalf(
			"expected false positive objection to be invalid; feedback=%q",
			result.Evaluation.Feedback,
		)
	}

	if result.Evaluation.Timely {
		t.Fatal("expected timely=false because there was no objection opportunity")
	}

	if result.Evaluation.Ruling != "overruled" {
		t.Fatalf("expected ruling overruled, got %q", result.Evaluation.Ruling)
	}

	requireNoMatchedOpportunity(t, result.Evaluation.MatchedOpportunityID)

	requireNormalizedObjectionType(
		t,
		result.Evaluation.NormalizedObjectionTypeID,
		"obj-hearsay",
	)

	if result.Evaluation.LegalAccuracyScore != 0 {
		t.Fatalf(
			"expected legal accuracy score 0 for false positive, got %.2f",
			result.Evaluation.LegalAccuracyScore,
		)
	}

	if result.Evaluation.StrategyScore != 0 {
		t.Fatalf(
			"expected strategy score 0 for false positive, got %.2f",
			result.Evaluation.StrategyScore,
		)
	}

	if !strings.Contains(result.Evaluation.Feedback, "no expected objection opportunity") {
		t.Fatalf(
			"expected feedback to mention no expected objection opportunity, got %q",
			result.Evaluation.Feedback,
		)
	}

	requireJudgeRuling(t, result.JudgeEvent, "Overruled.")

	requireCoachFeedbackContains(
		t,
		result.CoachEvent,
		"no expected objection opportunity",
	)
}

func TestCorrectPassOnCleanLine(t *testing.T) {
	t.Parallel()

	h := newSessionTestHarness(t)

	// Advance to line-hearsay-001, a clean line with no objection opportunity.
	advanceToLine(t, h, 1, "line-hearsay-001")

	result := submitTraineeAction(t, h, "pass", "Pass")

	if result.Action.ActionType != "pass" {
		t.Fatalf("expected action type pass, got %q", result.Action.ActionType)
	}

	if result.Action.RawText != "Pass" {
		t.Fatalf("expected raw text Pass, got %q", result.Action.RawText)
	}

	requireNoNormalizedObjectionType(t, result.Action.NormalizedObjectionTypeID)

	if !result.Evaluation.Valid {
		t.Fatalf(
			"expected pass on clean line to be valid; feedback=%q",
			result.Evaluation.Feedback,
		)
	}

	if !result.Evaluation.Timely {
		t.Fatal("expected timely=true because passing was correct on a clean line")
	}

	if result.Evaluation.Ruling != "no_ruling" {
		t.Fatalf("expected ruling no_ruling, got %q", result.Evaluation.Ruling)
	}

	requireNoMatchedOpportunity(t, result.Evaluation.MatchedOpportunityID)
	requireNoNormalizedObjectionType(t, result.Evaluation.NormalizedObjectionTypeID)

	if result.Evaluation.LegalAccuracyScore != 100 {
		t.Fatalf(
			"expected legal accuracy score 100 for correct pass, got %.2f",
			result.Evaluation.LegalAccuracyScore,
		)
	}

	if result.Evaluation.PhrasingScore != 100 {
		t.Fatalf(
			"expected phrasing score 100 for correct pass, got %.2f",
			result.Evaluation.PhrasingScore,
		)
	}

	if result.Evaluation.StrategyScore != 100 {
		t.Fatalf(
			"expected strategy score 100 for correct pass, got %.2f",
			result.Evaluation.StrategyScore,
		)
	}

	if !strings.Contains(result.Evaluation.Feedback, "Correct") {
		t.Fatalf(
			"expected feedback to mention Correct, got %q",
			result.Evaluation.Feedback,
		)
	}

	if !strings.Contains(result.Evaluation.Feedback, "no strong objection opportunity") {
		t.Fatalf(
			"expected feedback to mention no strong objection opportunity, got %q",
			result.Evaluation.Feedback,
		)
	}

	requireJudgeRuling(t, result.JudgeEvent, "No ruling.")

	requireCoachFeedbackContains(
		t,
		result.CoachEvent,
		"no strong objection opportunity",
	)
}

type sessionTestHarness struct {
	ctx     context.Context
	service *Service
	session *appdb.Session
}

func newSessionTestHarness(t *testing.T) *sessionTestHarness {
	t.Helper()

	ctx := context.Background()

	database := newTestDatabase(t)

	repo := NewGormRepository(database)
	service := NewService(repo, llm.NewNoopCoach())

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ScenarioID: "scenario-hearsay-001",
		Mode:       "spot_objection",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	return &sessionTestHarness{
		ctx:     ctx,
		service: service,
		session: session,
	}
}

func advanceToLine(
	t *testing.T,
	h *sessionTestHarness,
	steps int,
	expectedLineID string,
) *AdvanceSessionResult {
	t.Helper()

	var result *AdvanceSessionResult
	var err error

	for i := range steps {
		result, err = h.service.AdvanceSession(h.ctx, h.session.ID)
		if err != nil {
			t.Fatalf("advance session step %d: %v", i+1, err)
		}

		if result.Completed {
			t.Fatalf("session completed too early at step %d", i+1)
		}

		if result.Line == nil {
			t.Fatalf("expected line at step %d", i+1)
		}
	}

	if result.Line.ID != expectedLineID {
		t.Fatalf("expected %s, got %q", expectedLineID, result.Line.ID)
	}

	return result
}

func submitTraineeAction(
	t *testing.T,
	h *sessionTestHarness,
	actionType string,
	rawText string,
) *SubmitActionResult {
	t.Helper()

	result, err := h.service.SubmitAction(h.ctx, SubmitActionInput{
		SessionID:  h.session.ID,
		ActionType: actionType,
		RawText:    rawText,
	})
	if err != nil {
		t.Fatalf("submit trainee action %q %q: %v", actionType, rawText, err)
	}

	if result.Action == nil {
		t.Fatal("expected trainee action")
	}

	if result.Evaluation == nil {
		t.Fatal("expected action evaluation")
	}

	if result.JudgeEvent == nil {
		t.Fatal("expected judge event")
	}

	if result.CoachEvent == nil {
		t.Fatal("expected coach event")
	}

	return result
}

func requireNormalizedObjectionType(
	t *testing.T,
	actual *string,
	expected string,
) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected normalized objection type %q, got nil", expected)
	}

	if *actual != expected {
		t.Fatalf("expected normalized objection type %q, got %q", expected, *actual)
	}
}

func requireNoNormalizedObjectionType(t *testing.T, actual *string) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no normalized objection type, got %q", *actual)
	}
}

func requireMatchedOpportunity(
	t *testing.T,
	actual *string,
	expected string,
) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected matched opportunity %q, got nil", expected)
	}

	if *actual != expected {
		t.Fatalf("expected matched opportunity %q, got %q", expected, *actual)
	}
}

func requireNoMatchedOpportunity(t *testing.T, actual *string) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no matched opportunity, got %q", *actual)
	}
}

func requireJudgeRuling(
	t *testing.T,
	event *appdb.SessionEvent,
	expectedText string,
) {
	t.Helper()

	if event == nil {
		t.Fatal("expected judge event")
	}

	if event.EventType != "judge_ruling" {
		t.Fatalf("expected judge_ruling event, got %q", event.EventType)
	}

	if event.Text != expectedText {
		t.Fatalf("expected judge text %q, got %q", expectedText, event.Text)
	}
}

func requireCoachFeedbackContains(
	t *testing.T,
	event *appdb.SessionEvent,
	expectedSubstring string,
) {
	t.Helper()

	if event == nil {
		t.Fatal("expected coach event")
	}

	if event.EventType != "coach_feedback" {
		t.Fatalf("expected coach_feedback event, got %q", event.EventType)
	}

	if strings.TrimSpace(event.Text) == "" {
		t.Fatal("expected non-empty coach feedback")
	}

	if expectedSubstring != "" && !strings.Contains(event.Text, expectedSubstring) {
		t.Fatalf(
			"expected coach feedback to contain %q, got %q",
			expectedSubstring,
			event.Text,
		)
	}
}