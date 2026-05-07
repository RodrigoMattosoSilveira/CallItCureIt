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
	@echo "  make backend-tidy           Run go mod tidy"
	@echo "  make backend-test           Run backend tests"
	@echo "  make backend-run            Run backend API"
	@echo "  make check                  Run frontend and backend checks"

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

.PHONY: backend-tidy
backend-tidy:
	cd $(BACKEND_DIR) && go mod tidy

.PHONY: backend-test
backend-test:
	cd $(BACKEND_DIR) && go test ./...

.PHONY: backend-run
backend-run:
	cd $(BACKEND_DIR) && go run ./cmd/api

.PHONY: check
check: frontend-check backend-test