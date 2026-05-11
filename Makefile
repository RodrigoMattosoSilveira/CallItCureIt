SHELL := /bin/bash

FRONTEND_DIR := frontend
BACKEND_DIR := backend

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make frontend-install       Install frontend dependencies"
	@echo "  make frontend-dev           Run frontend dev server"
	@echo "  make frontend-typecheck     Run TypeScript checks"
	@echo "  make frontend-lint          Run ESLint"
	@echo "  make frontend-build         Build frontend"
	@echo "  make frontend-check         Run all frontend checks"
	@echo "  make frontend-e2e           Run frontend end-to-end tests"
	@echo "  make frontend-e2e-headed    Run frontend end-to-end tests in headed mode"
	@echo "  make frontend-e2e-ui        Run frontend end-to-end tests in UI mode"
	@echo "  make backend-tidy           Run go mod tidy"
	@echo "  make backend-test           Run backend tests"
	@echo "  make backend-run            Run backend API"
	@echo "  make check                  Run frontend and backend checks"
	@echo "  make check-with-e2e         Run all checks including end-to-end tests"
	@echo "  make openapi-lint           Lint OpenAPI specification"


.PHONY: frontend-install
frontend-install:
	cd $(FRONTEND_DIR) && npm install

.PHONY: frontend-dev
frontend-dev:
	cd $(FRONTEND_DIR) && npm run dev

.PHONY: frontend-typecheck
frontend-typecheck:
	cd $(FRONTEND_DIR) && npm run typecheck

.PHONY: frontend-lint
frontend-lint:
	cd $(FRONTEND_DIR) && npm run lint

.PHONY: frontend-build
frontend-build:
	cd $(FRONTEND_DIR) && npm run build

.PHONY: frontend-check
frontend-check:
	cd $(FRONTEND_DIR) && npm run check

.PHONY: frontend-e2e
frontend-e2e:
	cd $(FRONTEND_DIR) && npm run e2e


.PHONY: frontend-e2e-headed
frontend-e2e-headed:
	cd frontend && npm run e2e:headed

.PHONY: frontend-e2e-ui
frontend-e2e-ui:
	cd frontend && npm run e2e:ui

.PHONY: backend-tidy
backend-tidy:
	cd $(BACKEND_DIR) && go mod tidy

.PHONY: backend-dev
backend-dev:
    ./scripts/new_scores.sh
	./scripts/bedev.sh

.PHONY: backend-check
backend-test:
	cd $(BACKEND_DIR) && go test ./...

.PHONY: backend-run
backend-run:
	cd $(BACKEND_DIR) && go run ./cmd/api

.PHONY: check
check: frontend-check backend-test

.PHONY: check-with-e2e
check-with-e2e: backend-test frontend-check openapi-lint frontend-e2e

.PHONY: openapi-lint
openapi-lint:
	npx @redocly/cli lint backend/api/openapi.yaml