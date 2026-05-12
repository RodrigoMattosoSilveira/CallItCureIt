package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/scenarios"
	"CallItCureIt/backend/internal/sessions"

	"CallItCureIt/backend/internal/auth"
	"CallItCureIt/backend/internal/config"
	"CallItCureIt/backend/internal/llm"
)

func main() {
	cfg := config.Load()

	dbPath := cfg.DatabasePath
	port := cfg.Port

	database, err := db.ConnectSQLite(dbPath)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://192.168.2.154:5173",
		},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
	}))

	app.Get("/api/v1/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	authRepo := auth.NewGormRepository(database)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(app)

	scenarioRepo := scenarios.NewGormRepository(database)
	scenarioService := scenarios.NewService(scenarioRepo)
	scenarioHandler := scenarios.NewHandler(scenarioService)
	scenarioHandler.RegisterRoutes(app)

	adminScenarioRepo := scenarios.NewGormAdminRepository(database)
	adminScenarioService := scenarios.NewAdminService(adminScenarioRepo)
	adminScenarioHandler := scenarios.NewAdminHandler(adminScenarioService)
	adminScenarioHandler.RegisterRoutes(app, auth.RequireAdmin(authService))

	var coach llm.Coach = llm.NewNoopCoach()

	if cfg.LLMCoachingEnabled && cfg.OpenAIAPIKey != "" {
		coach = llm.NewOpenAICoach(
			cfg.OpenAIAPIKey,
			cfg.OpenAIModel,
			cfg.OpenAIBaseURL,
			cfg.OpenAITimeoutSeconds,
		)
	}

	sessionRepo := sessions.NewGormRepository(database)
	sessionService := sessions.NewService(sessionRepo, coach)
	sessionHandler := sessions.NewHandler(sessionService)
	sessionHandler.RegisterRoutes(app)

	log.Printf("API listening on :%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
