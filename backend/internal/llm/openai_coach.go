package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAICoach struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenAICoach(apiKey string, model string, baseURL string, timeoutSeconds int) *OpenAICoach {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}

	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAICoach{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

type responseRequest struct {
	Model       string       `json:"model"`
	Input       []inputItem  `json:"input"`
	Text        responseText `json:"text"`
	Temperature float64      `json:"temperature,omitempty"`
}

type inputItem struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responseText struct {
	Format responseFormat `json:"format"`
}

type responseFormat struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Schema any    `json:"schema,omitempty"`
	Strict bool   `json:"strict,omitempty"`
}

type coachingResponse struct {
	Feedback string `json:"feedback"`
}

func (c *OpenAICoach) EnhanceFeedback(ctx context.Context, input CoachingInput) (string, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return "", errors.New("openai api key is empty")
	}

	systemPrompt := strings.TrimSpace(`
You are a litigation training coach.

You do not decide whether the objection was correct.
The deterministic legal engine has already decided validity, timeliness, ruling, and scores.

Your job:
- Explain the result clearly and briefly.
- Preserve the deterministic ruling and correctness.
- Do not invent legal rules, citations, facts, or exceptions.
- Do not give real legal advice.
- Keep the tone practical, courtroom-oriented, and instructional.
- Include a better courtroom phrase when useful.
- Return only JSON matching the requested schema.
`)

	userPrompt := buildCoachingPrompt(input)

	payload := responseRequest{
		Model: c.model,
		Input: []inputItem{
			{
				Role: "system",
				Content: []contentBlock{
					{
						Type: "input_text",
						Text: systemPrompt,
					},
				},
			},
			{
				Role: "user",
				Content: []contentBlock{
					{
						Type: "input_text",
						Text: userPrompt,
					},
				},
			},
		},
		Text: responseText{
			Format: responseFormat{
				Type:   "json_schema",
				Name:   "coaching_feedback",
				Strict: true,
				Schema: map[string]any{
					"type": "object",
					"additionalProperties": false,
					"required": []string{
						"feedback",
					},
					"properties": map[string]any{
						"feedback": map[string]any{
							"type":      "string",
							"minLength": 1,
							"maxLength": 1200,
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/responses",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call openai responses api: %w", err)
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("read openai response: %w", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai response status %d: %s", resp.StatusCode, string(respBody))
	}

	outputText, err := extractOutputText(respBody)
	if err != nil {
		return "", err
	}

	var parsed coachingResponse
	if err := json.Unmarshal([]byte(outputText), &parsed); err != nil {
		return "", fmt.Errorf("parse coaching json: %w; raw=%s", err, outputText)
	}

	feedback := strings.TrimSpace(parsed.Feedback)
	if feedback == "" {
		return "", errors.New("empty coaching feedback")
	}

	return feedback, nil
}

func buildCoachingPrompt(input CoachingInput) string {
	return fmt.Sprintf(`Create improved coaching feedback for this objection training event.

Scenario ID: %s

Transcript line:
- Speaker: %s
- Kind: %s
- Text: %s

Trainee action:
%s

Deterministic evaluation:
- Valid: %t
- Timely: %t
- Ruling: %s
- Normalized objection type ID: %s
- Matched opportunity ID: %s
- Legal accuracy score: %.2f
- Phrasing score: %.2f
- Strategy score: %.2f

Expected opportunity:
- Expected phrase: %s
- Explanation: %s

Current deterministic feedback:
%s

Write concise coaching feedback for a lawyer in training.
Do not contradict the ruling or scores.
Do not add citations unless they appear in the input.
`,
		input.ScenarioID,
		input.SpeakerName,
		input.LineKind,
		input.LineText,
		input.TraineeAction,
		input.Valid,
		input.Timely,
		input.Ruling,
		input.NormalizedObjectionTypeID,
		input.MatchedOpportunityID,
		input.LegalAccuracyScore,
		input.PhrasingScore,
		input.StrategyScore,
		input.ExpectedPhrase,
		input.ExpectedObjectionExplanation,
		input.DeterministicFeedback,
	)
}

func extractOutputText(body []byte) (string, error) {
	var raw struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}

	if strings.TrimSpace(raw.OutputText) != "" {
		return raw.OutputText, nil
	}

	for _, item := range raw.Output {
		for _, content := range item.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text, nil
			}
		}
	}

	return "", errors.New("openai response did not contain output text")
}