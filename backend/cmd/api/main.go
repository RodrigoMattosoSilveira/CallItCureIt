package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"CallItCureIt/backend/internal/db"
	"CallItCureIt/backend/internal/scenarios"
)

func main() {
	dbPath := getEnv("DATABASE_PATH", "data/app.db")
	port := getEnv("PORT", "8080")

	database, err := db.ConnectSQLite(dbPath)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
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

	scenarioRepo := scenarios.NewGormRepository(database)
	scenarioService := scenarios.NewService(scenarioRepo)
	scenarioHandler := scenarios.NewHandler(scenarioService)
	scenarioHandler.RegisterRoutes(app)

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
