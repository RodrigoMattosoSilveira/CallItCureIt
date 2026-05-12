package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"CallItCureIt/backend/internal/admin"
	"CallItCureIt/backend/internal/auth"
	"CallItCureIt/backend/internal/config"
	"CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/llm"
	"CallItCureIt/backend/internal/scenarios"
	"CallItCureIt/backend/internal/sessions"
)

func main() {
	cfg := config.Load()

	database, err := db.ConnectSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: parseAllowedOrigins(getEnv(
			"CORS_ALLOW_ORIGINS",
			"http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173",
		)),
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

	// Auth
	authRepo := auth.NewGormRepository(database)
	authService := auth.NewService(authRepo, cfg)

	if err := authService.EnsureDevAdmin(context.Background()); err != nil {
		log.Fatalf("failed to seed dev admin: %v", err)
	}

	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(app)

	// Public scenario routes
	scenarioRepo := scenarios.NewGormRepository(database)
	scenarioService := scenarios.NewService(scenarioRepo)
	scenarioHandler := scenarios.NewHandler(scenarioService)
	scenarioHandler.RegisterRoutes(app)

	adminGroup := app.Group(
		"/api/v1/admin",
		auth.RequireAuth(authService),
		auth.RequireAdmin(),
	)

	adminScenarioHandler := admin.NewScenarioHandler(scenarioService)
	adminScenarioHandler.RegisterRoutes(adminGroup)	

	// Session routes
	sessionRepo := sessions.NewGormRepository(database)

	var coach llm.Coach = llm.NewNoopCoach()

	if cfg.LLMCoachingEnabled {
		coach = llm.NewOpenAICoach(
			cfg.OpenAIAPIKey,
			cfg.OpenAIModel,
			cfg.OpenAIBaseURL,
			cfg.OpenAITimeoutSeconds,
		)
	}

	sessionService := sessions.NewService(sessionRepo, coach)
	sessionHandler := sessions.NewHandler(sessionService)
	sessionHandler.RegisterRoutes(app)

	// Protected admin route group.
	//
	// Keep this group here so future admin handlers can be registered like:
	//
	// admin := app.Group(
	// 	"/api/v1/admin",
	// 	auth.RequireAuth(authService),
	// 	auth.RequireAdmin(),
	// )
	//
	// admin.Get("/scenarios", adminScenarioHandler.ListScenarios)
	//
	// Do not register /api/v1/admin/scenarios here unless you already have an
	// admin scenario handler implemented.

	log.Printf("API listening on :%s", cfg.Port)
	log.Printf("Database path: %s", cfg.DatabasePath)
	log.Printf("Dev admin seed enabled: %v", cfg.DevSeedAdmin)
	log.Printf("Dev admin email: %s", cfg.DevAdminEmail)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func parseAllowedOrigins(value string) []string {
	parts := strings.Split(value, ",")

	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}