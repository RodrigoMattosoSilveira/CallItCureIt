SHELL := /usr/bin/env bash

# -----------------------------------------------------------------------------
# Call It Cure It - Project Makefile
# -----------------------------------------------------------------------------
# Usage examples:
#   make init-dev
#   make dev-backend
#   make dev-frontend
#   make check
#   make docker-prod-build
#   make docker-prod-up
#   make prod-smoke BASE_URL=https://app.callitcureit.com
# -----------------------------------------------------------------------------

APP_NAME := call-it-cure-it

BACKEND_DIR := backend
FRONTEND_DIR := frontend
SCRIPTS_DIR := scripts

DEV_BACKEND_ENV := $(BACKEND_DIR)/.env
DEV_FRONTEND_ENV := $(FRONTEND_DIR)/.env
PROD_ENV ?= .env.production

LOCAL_DB := $(BACKEND_DIR)/data/app.db
LOCAL_DB_DIR := $(BACKEND_DIR)/data

PROD_COMPOSE := docker-compose.prod.yml
DEV_COMPOSE := docker-compose.dev.yml

BASE_URL ?= http://localhost
API_BASE_URL ?= $(BASE_URL)/api/v1

.PHONY: help
help:
	@echo ""
	@echo "Call It Cure It - Make targets"
	@echo ""
	@echo "Local setup:"
	@echo "  make init-dev              Create/update backend/.env and frontend/.env"
	@echo "  make db-init               Create local SQLite DB and apply migrations"
	@echo "  make db-reset              Delete and recreate local SQLite DB"
	@echo ""
	@echo "Local development:"
	@echo "  make dev-backend           Run backend locally"
	@echo "  make dev-frontend          Run frontend locally"
	@echo "  make dev                   Print commands for running both dev servers"
	@echo ""
	@echo "Checks:"
	@echo "  make backend-check         Run backend tests"
	@echo "  make frontend-check        Run frontend checks"
	@echo "  make check                 Run backend + frontend checks"
	@echo ""
	@echo "Smoke tests:"
	@echo "  make local-smoke           Test local Vite proxy/API"
	@echo "  make prod-smoke            Run production smoke test"
	@echo ""
	@echo "Docker production:"
	@echo "  make init-prod             Create/update .env.production"
	@echo "  make docker-prod-build     Build production containers"
	@echo "  make docker-prod-up        Start production stack"
	@echo "  make docker-prod-down      Stop production stack"
	@echo "  make docker-prod-ps        Show production containers"
	@echo "  make docker-prod-logs      Follow production logs"
	@echo ""
	@echo "Maintenance:"
	@echo "  make prod-backup           Backup production SQLite DB"
	@echo "  make docker-clean          Clean unused Docker cache/images/containers"
	@echo ""

# -----------------------------------------------------------------------------
# Initialization
# -----------------------------------------------------------------------------

.PHONY: init-dev
init-dev:
	@chmod +x $(SCRIPTS_DIR)/init-dev-env.sh
	@./$(SCRIPTS_DIR)/init-dev-env.sh

.PHONY: init-prod
init-prod:
	@chmod +x $(SCRIPTS_DIR)/init-prod-env.sh
	@./$(SCRIPTS_DIR)/init-prod-env.sh $(PROD_ENV)

.PHONY: chmod-scripts
chmod-scripts:
	@chmod +x $(SCRIPTS_DIR)/*.sh

# -----------------------------------------------------------------------------
# Local SQLite database
# -----------------------------------------------------------------------------

.PHONY: db-init
db-init:
	@mkdir -p $(LOCAL_DB_DIR)
	@echo "Applying migrations to $(LOCAL_DB)..."
	@for migration in $(BACKEND_DIR)/migrations/*.up.sql; do \
		echo "Applying $$migration"; \
		sqlite3 $(LOCAL_DB) < "$$migration"; \
	done
	@echo "Done."
	@sqlite3 $(LOCAL_DB) ".tables"

.PHONY: db-reset
db-reset:
	@echo "Removing $(LOCAL_DB)..."
	@rm -f $(LOCAL_DB) $(LOCAL_DB)-shm $(LOCAL_DB)-wal
	@$(MAKE) db-init

.PHONY: db-tables
db-tables:
	@sqlite3 $(LOCAL_DB) ".tables"

.PHONY: db-schema
db-schema:
	@sqlite3 $(LOCAL_DB) ".schema"

.PHONY: db-reset-sessions
db-reset-sessions:
	@sqlite3 $(LOCAL_DB) "\
		DELETE FROM session_scores; \
		DELETE FROM action_evaluations; \
		DELETE FROM trainee_actions; \
		DELETE FROM session_events; \
		DELETE FROM sessions; \
	"
	@echo "Session/test data cleared."

.PHONY: db-reset-admin
db-reset-admin:
	@sqlite3 $(LOCAL_DB) "DELETE FROM users WHERE email = 'admin@example.com';"
	@echo "Local admin user cleared. Restart backend to reseed."

# -----------------------------------------------------------------------------
# Local development
# -----------------------------------------------------------------------------

.PHONY: dev-backend
dev-backend:
	@chmod +x $(SCRIPTS_DIR)/dev-backend.sh
	@./$(SCRIPTS_DIR)/dev-backend.sh

.PHONY: dev-frontend
dev-frontend:
	@chmod +x $(SCRIPTS_DIR)/dev-frontend.sh
	@./$(SCRIPTS_DIR)/dev-frontend.sh

.PHONY: dev
dev:
	@echo ""
	@echo "Run these in two separate terminals:"
	@echo ""
	@echo "  make dev-backend"
	@echo "  make dev-frontend"
	@echo ""
	@echo "Then open:"
	@echo "  http://localhost:5173"
	@echo ""

# -----------------------------------------------------------------------------
# Checks
# -----------------------------------------------------------------------------

.PHONY: backend-check
backend-check:
	@cd $(BACKEND_DIR) && go test ./...

.PHONY: frontend-check
frontend-check:
	@cd $(FRONTEND_DIR) && npm run check

.PHONY: check
check: backend-check frontend-check

.PHONY: backend-tidy
backend-tidy:
	@cd $(BACKEND_DIR) && go mod tidy

.PHONY: frontend-install
frontend-install:
	@cd $(FRONTEND_DIR) && npm install

# -----------------------------------------------------------------------------
# Local smoke tests
# -----------------------------------------------------------------------------

.PHONY: local-smoke
local-smoke:
	@echo "Testing local backend through Vite proxy..."
	@curl -fsS http://localhost:5173/api/v1/healthz | grep -q "ok"
	@echo "Health OK"
	@curl -fsS http://localhost:5173/api/v1/scenarios | grep -q "data"
	@echo "Scenarios OK"
	@echo "Local smoke test passed."

.PHONY: local-api-smoke
local-api-smoke:
	@echo "Testing local backend directly..."
	@curl -fsS http://localhost:8080/api/v1/healthz | grep -q "ok"
	@echo "Health OK"
	@curl -fsS http://localhost:8080/api/v1/scenarios | grep -q "data"
	@echo "Scenarios OK"
	@echo "Local API smoke test passed."

.PHONY: local-login-test
local-login-test:
	@curl -s -X POST http://localhost:5173/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@example.com","password":"admin123"}' | jq

# -----------------------------------------------------------------------------
# Docker development
# -----------------------------------------------------------------------------

.PHONY: docker-dev-build
docker-dev-build:
	@docker compose -f $(DEV_COMPOSE) build

.PHONY: docker-dev-up
docker-dev-up:
	@docker compose -f $(DEV_COMPOSE) up -d

.PHONY: docker-dev-down
docker-dev-down:
	@docker compose -f $(DEV_COMPOSE) down

.PHONY: docker-dev-logs
docker-dev-logs:
	@docker compose -f $(DEV_COMPOSE) logs -f --tail=200

# -----------------------------------------------------------------------------
# Docker production
# -----------------------------------------------------------------------------

.PHONY: docker-prod-build
docker-prod-build:
	@chmod +x $(SCRIPTS_DIR)/prod-build.sh
	@./$(SCRIPTS_DIR)/prod-build.sh $(PROD_ENV)

.PHONY: docker-prod-build-plain
docker-prod-build-plain:
	@DOCKER_BUILDKIT=1 BUILDX_NO_DEFAULT_ATTESTATIONS=1 docker compose \
		--env-file $(PROD_ENV) \
		-f $(PROD_COMPOSE) \
		build --progress=plain

.PHONY: docker-prod-up
docker-prod-up:
	@chmod +x $(SCRIPTS_DIR)/prod-up.sh
	@./$(SCRIPTS_DIR)/prod-up.sh $(PROD_ENV)

.PHONY: docker-prod-down
docker-prod-down:
	@chmod +x $(SCRIPTS_DIR)/prod-down.sh
	@./$(SCRIPTS_DIR)/prod-down.sh $(PROD_ENV)

.PHONY: docker-prod-ps
docker-prod-ps:
	@docker compose --env-file $(PROD_ENV) -f $(PROD_COMPOSE) ps -a

.PHONY: docker-prod-logs
docker-prod-logs:
	@chmod +x $(SCRIPTS_DIR)/prod-logs.sh
	@./$(SCRIPTS_DIR)/prod-logs.sh $(PROD_ENV)

.PHONY: docker-prod-restart
docker-prod-restart:
	@docker compose --env-file $(PROD_ENV) -f $(PROD_COMPOSE) up -d

.PHONY: docker-prod-restart-caddy
docker-prod-restart-caddy:
	@docker compose --env-file $(PROD_ENV) -f $(PROD_COMPOSE) restart caddy

# -----------------------------------------------------------------------------
# Production health and smoke tests
# -----------------------------------------------------------------------------

.PHONY: prod-backend-health
prod-backend-health:
	@docker exec -it callitcureit-backend curl -i http://localhost:8080/api/v1/healthz

.PHONY: prod-caddy-health
prod-caddy-health:
	@curl -i http://localhost/api/v1/healthz

.PHONY: prod-public-health
prod-public-health:
	@curl -i $(BASE_URL)/api/v1/healthz

.PHONY: prod-smoke
prod-smoke:
	@chmod +x $(SCRIPTS_DIR)/prod-smoke-test.sh
	@BASE_URL=$(BASE_URL) ./$(SCRIPTS_DIR)/prod-smoke-test.sh

.PHONY: prod-cert-check
prod-cert-check:
	@echo | openssl s_client -connect app.callitcureit.com:443 -servername app.callitcureit.com -showcerts 2>/dev/null \
		| openssl x509 -noout -subject -issuer -dates -ext subjectAltName

# -----------------------------------------------------------------------------
# Production admin tests
# -----------------------------------------------------------------------------

.PHONY: prod-login-test
prod-login-test:
	@if [[ -z "$$ADMIN_EMAIL" || -z "$$ADMIN_PASSWORD" ]]; then \
		echo "Usage: ADMIN_EMAIL=... ADMIN_PASSWORD=... BASE_URL=https://app.callitcureit.com make prod-login-test"; \
		exit 1; \
	fi
	@curl -s -X POST $(BASE_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d "{\"email\":\"$$ADMIN_EMAIL\",\"password\":\"$$ADMIN_PASSWORD\"}" | jq

.PHONY: prod-admin-scenarios-test
prod-admin-scenarios-test:
	@if [[ -z "$$ADMIN_EMAIL" || -z "$$ADMIN_PASSWORD" ]]; then \
		echo "Usage: ADMIN_EMAIL=... ADMIN_PASSWORD=... BASE_URL=https://app.callitcureit.com make prod-admin-scenarios-test"; \
		exit 1; \
	fi
	@TOKEN=$$(curl -s -X POST $(BASE_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d "{\"email\":\"$$ADMIN_EMAIL\",\"password\":\"$$ADMIN_PASSWORD\"}" | jq -r '.data.token'); \
	curl -s $(BASE_URL)/api/v1/admin/scenarios \
		-H "Authorization: Bearer $$TOKEN" | jq

# -----------------------------------------------------------------------------
# Backups and maintenance
# -----------------------------------------------------------------------------

.PHONY: prod-backup
prod-backup:
	@chmod +x $(SCRIPTS_DIR)/prod-backup-sqlite.sh
	@./$(SCRIPTS_DIR)/prod-backup-sqlite.sh

.PHONY: docker-clean
docker-clean:
	@docker builder prune -f
	@docker image prune -f
	@docker container prune -f

.PHONY: docker-df
docker-df:
	@docker system df

.PHONY: disk
disk:
	@df -h
	@echo ""
	@docker system df || true

# -----------------------------------------------------------------------------
# Git / deployment helpers
# -----------------------------------------------------------------------------

.PHONY: pull
pull:
	@git pull

.PHONY: deploy-prod
deploy-prod: pull docker-prod-build docker-prod-up prod-smoke

.PHONY: deploy-prod-no-smoke
deploy-prod-no-smoke: pull docker-prod-build docker-prod-up