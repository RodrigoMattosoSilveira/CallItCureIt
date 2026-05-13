SHELL := /bin/bash

BACKEND_DIR := backend
FRONTEND_DIR := frontend

.PHONY: help
help:
	@echo "Available commands:"
	@echo
	@echo "Database:"
	@echo "  make db-migrate          Apply all SQLite migrations"
	@echo "  make db-reset            Delete and recreate local SQLite database"
	@echo
	@echo "Development:"
	@echo "  make dev-backend         Run backend API"
	@echo "  make dev-frontend        Run frontend dev server"
	@echo
	@echo "Backend:"
	@echo "  make backend-tidy        Run go mod tidy"
	@echo "  make backend-test        Run backend tests"
	@echo "  make backend-test-v      Run backend tests verbose"
	@echo
	@echo "Frontend:"
	@echo "  make frontend-install    Install frontend dependencies"
	@echo "  make frontend-check      Typecheck, lint, and build frontend"
	@echo "  make frontend-e2e        Run Playwright e2e tests"
	@echo "  make frontend-e2e-headed Run Playwright e2e tests headed"
	@echo "  make frontend-e2e-ui     Run Playwright UI"
	@echo
	@echo "Quality:"
	@echo "  make openapi-lint        Lint OpenAPI spec"
	@echo "  make check               Run backend, frontend, and OpenAPI checks"
	@echo "  make check-with-e2e      Run check plus e2e tests"
	@echo
	@echo "Smoke tests:"
	@echo "  make test-hearsay-flow   Run curl backend hearsay flow"

.PHONY: db-migrate
db-migrate:
	./scripts/db-migrate.sh

.PHONY: db-reset
db-reset:
	./scripts/db-reset.sh

.PHONY: dev-backend
dev-backend:
	./scripts/dev-backend.sh

.PHONY: dev-frontend
dev-frontend:
	./scripts/dev-frontend.sh

.PHONY: backend-tidy
backend-tidy:
	cd $(BACKEND_DIR) && go mod tidy

.PHONY: backend-test
backend-test:
	cd $(BACKEND_DIR) && go test ./...

.PHONY: backend-test-v
backend-test-v:
	cd $(BACKEND_DIR) && go test ./... -v

.PHONY: frontend-install
frontend-install:
	cd $(FRONTEND_DIR) && npm install

.PHONY: frontend-check
frontend-check:
	cd $(FRONTEND_DIR) && npm run check

.PHONY: frontend-e2e
frontend-e2e:
	cd $(FRONTEND_DIR) && npm run e2e

.PHONY: frontend-e2e-headed
frontend-e2e-headed:
	cd $(FRONTEND_DIR) && npm run e2e:headed

.PHONY: frontend-e2e-ui
frontend-e2e-ui:
	cd $(FRONTEND_DIR) && npm run e2e:ui

.PHONY: openapi-lint
openapi-lint:
	npx @redocly/cli lint backend/api/openapi.yaml

.PHONY: check
check:
	./scripts/check.sh

.PHONY: check-with-e2e
check-with-e2e: check frontend-e2e

.PHONY: test-hearsay-flow
test-hearsay-flow:
	./scripts/test-hearsay-flow.sh

.PHONY: docker-dev-up
docker-dev-up:
	docker compose -f docker-compose.dev.yml up --build

.PHONY: docker-dev-down
docker-dev-down:
	docker compose -f docker-compose.dev.yml down

.PHONY: docker-dev-logs
docker-dev-logs:
	docker compose -f docker-compose.dev.yml logs -f

.PHONY: docker-prod-up
docker-prod-up:
	docker compose up --build -d

.PHONY: docker-prod-down
docker-prod-down:
	docker compose down

.PHONY: docker-prod-logs
docker-prod-logs:
	docker compose logs -f

.PHONY: docker-migrate
docker-migrate:
	./scripts/docker-migrate.sh

.PHONY: docker-ensure-admin-exists
docker-ensure-admin-exists:
	./scripts/ensure-admin-exists.sh

.PHONY: docker-prod-reset
docker-prod-reset:
	docker compose down -v
	docker compose up --build -d backend
	./scripts/docker-migrate.sh
	./scripts/ensure-admin-exists.sh
	docker compose up -d