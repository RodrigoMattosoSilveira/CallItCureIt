package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	DatabasePath         string
	LLMCoachingEnabled   bool
	OpenAIAPIKey         string
	OpenAIModel          string
	OpenAIBaseURL        string
	OpenAITimeoutSeconds int
}

func Load() Config {
	return Config{
		Port:                 getEnv("PORT", "8080"),
		DatabasePath:         getEnv("DATABASE_PATH", "data/app.db"),
		LLMCoachingEnabled:   getEnvBool("LLM_COACHING_ENABLED", false),
		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:          getEnv("OPENAI_MODEL", "gpt-5.1-mini"),
		OpenAIBaseURL:        getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAITimeoutSeconds: getEnvInt("OPENAI_TIMEOUT_SECONDS", 20),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}