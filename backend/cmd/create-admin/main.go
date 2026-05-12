package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"CallItCureIt/backend/internal/auth"
	"CallItCureIt/backend/internal/config"
	"CallItCureIt/backend/internal/db"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	fullName := os.Getenv("ADMIN_FULL_NAME")

	if fullName == "" {
		fullName = "Admin User"
	}

	if email == "" || password == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD are required")
	}

	database, err := db.ConnectSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	repo := auth.NewGormRepository(database)
	service := auth.NewService(repo, cfg.JWTSecret)

	user, err := service.CreateUser(ctx, auth.CreateUserInput{
		Email:    email,
		Password: password,
		FullName: fullName,
		Role:     "admin",
	})
	if err != nil {
		log.Fatalf("create admin user: %v", err)
	}

	fmt.Printf("Created admin user: %s <%s>\n", user.FullName, user.Email)
}