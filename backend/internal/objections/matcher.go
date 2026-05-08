package objections

import (
	"strings"
	"unicode"
)

type NormalizedObjection struct {
	ObjectionTypeID string
	Code            string
	Confidence      float64
	Matched         bool
}

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Normalize(rawText string) NormalizedObjection {
	text := normalizeText(rawText)

	rules := []struct {
		code string
		id   string
		keys []string
	}{
		{
			code: "hearsay",
			id:   "obj-hearsay",
			keys: []string{
				"hearsay",
				"out of court statement",
				"out-of-court statement",
				"offered for the truth",
			},
		},
		{
			code: "relevance",
			id:   "obj-relevance",
			keys: []string{
				"relevance",
				"irrelevant",
				"not relevant",
			},
		},
		{
			code: "foundation",
			id:   "obj-foundation",
			keys: []string{
				"foundation",
				"lack of foundation",
				"no foundation",
				"proper foundation",
			},
		},
		{
			code: "leading",
			id:   "obj-leading",
			keys: []string{
				"leading",
				"leading question",
			},
		},
		{
			code: "speculation",
			id:   "obj-speculation",
			keys: []string{
				"speculation",
				"speculative",
				"calls for speculation",
				"personal knowledge",
			},
		},
		{
			code: "asked_and_answered",
			id:   "obj-asked-answered",
			keys: []string{
				"asked and answered",
				"already answered",
			},
		},
		{
			code: "argumentative",
			id:   "obj-argumentative",
			keys: []string{
				"argumentative",
				"arguing with the witness",
			},
		},
		{
			code: "compound",
			id:   "obj-compound",
			keys: []string{
				"compound",
				"compound question",
				"multiple questions",
			},
		},
	}

	for _, rule := range rules {
		for _, key := range rule.keys {
			if strings.Contains(text, normalizeText(key)) {
				return NormalizedObjection{
					ObjectionTypeID: rule.id,
					Code:            rule.code,
					Confidence:      1.0,
					Matched:         true,
				}
			}
		}
	}

	return NormalizedObjection{
		Confidence: 0,
		Matched:    false,
	}
}

func normalizeText(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))

	var builder strings.Builder
	lastWasSpace := false

	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastWasSpace = false
			continue
		}

		if !lastWasSpace {
			builder.WriteRune(' ')
			lastWasSpace = true
		}
	}

	return strings.TrimSpace(builder.String())
}