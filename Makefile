SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

SERVER_ROOT ?= /opt/CallItCureIt
EDGE_DIR := $(SERVER_ROOT)/edge
LAN_HOST ?= 192.168.2.154
ENV ?= development

ifeq ($(ENV),development)
  ENV_DIR := $(SERVER_ROOT)/development
  ENV_FILE := .env.development
  COMPOSE_PROJECT := callitcureit-dev
  BRANCH := development
  DOMAIN := dev.callitcureit.com
endif
ifeq ($(ENV),test)
  ENV_DIR := $(SERVER_ROOT)/test
  ENV_FILE := .env.test
  COMPOSE_PROJECT := callitcureit-tst
  BRANCH := test
  DOMAIN := tst.callitcureit.com
endif
ifeq ($(ENV),production)
  ENV_DIR := $(SERVER_ROOT)/production
  ENV_FILE := .env.production
  COMPOSE_PROJECT := callitcureit-prd
  BRANCH := production
  DOMAIN := app.callitcureit.com
endif

PROXY_DIR := $(SERVER_ROOT)/reverse-proxy

.PHONY: help
help:
	@echo "Call It Cure It Makefile"
	@echo "Local: make doctor local-init-env local-db-init local-check local-backend local-frontend local-smoke local-admin-test"
	@echo "Server: make server-pull/server-build/server-up/server-smoke ENV=development|test|production"
	@echo "Aliases: make server-dev-up server-test-up server-prod-up"

.PHONY: doctor
doctor:
	@command -v go >/dev/null || (echo "go not found" && exit 1)
	@command -v node >/dev/null || (echo "node not found" && exit 1)
	@command -v npm >/dev/null || (echo "npm not found" && exit 1)
	@command -v sqlite3 >/dev/null || (echo "sqlite3 not found" && exit 1)
	@command -v docker >/dev/null || (echo "docker not found" && exit 1)
	@docker compose version >/dev/null || (echo "docker compose not found" && exit 1)
	@command -v jq >/dev/null || (echo "jq not found" && exit 1)
	@command -v git >/dev/null || (echo "git not found" && exit 1)
	@echo "All required tools are available."

.PHONY: check-repo
check-repo:
	@test -f backend/Dockerfile
	@test -f backend/docker-entrypoint.sh
	@test -f frontend/Dockerfile
	@test -f frontend/nginx.conf
	@test -f docker-compose.server.yml
	@test -f reverse-proxy/docker-compose.proxy.yml
	@test -f reverse-proxy/Caddyfile
	@test -f scripts/init-dev-env.sh
	@test -f scripts/init-server-env.sh
	@if [ -d backend/cmd/create-admin ] || [ -d backend/cmd/create-admin.disabled ]; then echo "Remove obsolete backend/cmd/create-admin"; exit 1; fi
	@echo "Repository structure looks good."

.PHONY: local-init-env
local-init-env:
	chmod +x scripts/init-dev-env.sh
	./scripts/init-dev-env.sh

.PHONY: local-db-init
local-db-init:
	mkdir -p backend/data
	for migration in backend/migrations/*.up.sql; do echo "Applying $$migration"; sqlite3 backend/data/app.db < "$$migration"; done
	sqlite3 backend/data/app.db ".tables"

.PHONY: local-db-reset
local-db-reset:
	sqlite3 backend/data/app.db "DELETE FROM session_scores; DELETE FROM action_evaluations; DELETE FROM trainee_actions; DELETE FROM session_events; DELETE FROM sessions;"

.PHONY: local-admin-reset
local-admin-reset:
	sqlite3 backend/data/app.db "DELETE FROM users WHERE email = 'admin@example.com';"

.PHONY: backend-check
backend-check:
	cd backend && go mod tidy && go test ./...

.PHONY: frontend-check
frontend-check:
	cd frontend && npm install && npm run check

.PHONY: local-check
local-check: check-repo backend-check frontend-check

.PHONY: local-backend
local-backend:
	chmod +x scripts/dev-backend.sh
	./scripts/dev-backend.sh

.PHONY: local-frontend
local-frontend:
	chmod +x scripts/dev-frontend.sh
	./scripts/dev-frontend.sh

.PHONY: local-smoke
local-smoke:
	curl -fsS http://localhost:8080/api/v1/healthz >/dev/null
	curl -fsS http://localhost:5173/api/v1/healthz >/dev/null
	curl -fsS http://localhost:5173/api/v1/scenarios >/dev/null
	@echo "Local smoke tests passed."

.PHONY: local-lan-smoke
local-lan-smoke:
	curl -fsS http://$(LAN_HOST):5173/api/v1/healthz >/dev/null

.PHONY: local-login-test
local-login-test:
	curl -fsS -X POST http://localhost:5173/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"admin123"}' | jq .

.PHONY: local-admin-test
local-admin-test:
	TOKEN=$$(curl -fsS -X POST http://localhost:5173/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"admin123"}' | jq -r '.data.token'); \
	curl -fsS http://localhost:5173/api/v1/admin/scenarios -H "Authorization: Bearer $$TOKEN" | jq .

.PHONY: proxy-down
proxy-down:
	cd $(PROXY_DIR) && docker compose -f docker-compose.proxy.yml down

.PHONY: proxy-logs
proxy-logs:
	docker logs -f --tail=200 callitcureit-caddy

.PHONY: server-init-env
server-init-env:
	chmod +x scripts/init-server-env.sh
	./scripts/init-server-env.sh $(ENV)

.PHONY: server-pull
server-pull:
	cd $(ENV_DIR) && git checkout $(BRANCH) && git pull

.PHONY: server-build
server-build:
	cd $(ENV_DIR) && BUILDX_NO_DEFAULT_ATTESTATIONS=1 docker compose \
		--progress plain \
		-p $(COMPOSE_PROJECT) \
		--env-file $(ENV_FILE) \
		-f docker-compose.server.yml \
		build

.PHONY: server-up
server-up:
	cd $(ENV_DIR) && docker compose \
		-p $(COMPOSE_PROJECT) \
		--env-file $(ENV_FILE) \
		-f docker-compose.server.yml \
		up -d
		
.PHONY: server-down
server-down:
	cd $(ENV_DIR) && docker compose -p $(COMPOSE_PROJECT) --env-file $(ENV_FILE) -f docker-compose.server.yml down

.PHONY: server-ps
server-ps:
	cd $(ENV_DIR) && docker compose -p $(COMPOSE_PROJECT) --env-file $(ENV_FILE) -f docker-compose.server.yml ps -a

.PHONY: server-logs
server-logs:
	cd $(ENV_DIR) && docker compose -p $(COMPOSE_PROJECT) --env-file $(ENV_FILE) -f docker-compose.server.yml logs -f --tail=200

.PHONY: server-backend-health
server-backend-health:
	docker exec -it $(COMPOSE_PROJECT)-backend curl -i http://localhost:8080/api/v1/healthz

.PHONY: server-smoke
server-smoke:
	curl -fsS https://$(DOMAIN)/api/v1/healthz >/dev/null
	curl -fsS https://$(DOMAIN)/api/v1/scenarios >/dev/null
	@echo "$(DOMAIN) smoke tests passed."

.PHONY: server-admin-test
server-admin-test:
	cd $(ENV_DIR) && \
	ADMIN_EMAIL=$$(grep '^DEV_ADMIN_EMAIL=' $(ENV_FILE) | cut -d '=' -f2-); \
	ADMIN_PASSWORD=$$(grep '^DEV_ADMIN_PASSWORD=' $(ENV_FILE) | cut -d '=' -f2-); \
	TOKEN=$$(curl -fsS -X POST https://$(DOMAIN)/api/v1/auth/login -H "Content-Type: application/json" -d "$$(printf '{"email":"%s","password":"%s"}' "$$ADMIN_EMAIL" "$$ADMIN_PASSWORD")" | jq -r '.data.token'); \
	curl -fsS https://$(DOMAIN)/api/v1/admin/scenarios -H "Authorization: Bearer $$TOKEN" | jq .

.PHONY: server-dns-check
server-dns-check:
	dig $(DOMAIN)

.PHONY: server-cert-check
server-cert-check:
	echo | openssl s_client -connect $(DOMAIN):443 -servername $(DOMAIN) -showcerts 2>/dev/null | openssl x509 -noout -subject -issuer -dates -ext subjectAltName

.PHONY: server-backup
server-backup:
	mkdir -p $(ENV_DIR)/backups
	TIMESTAMP=$$(date +%Y%m%d-%H%M%S); CONTAINER="$(COMPOSE_PROJECT)-backend"; \
	docker exec $$CONTAINER sqlite3 /app/data/app.db ".backup '/tmp/app-backup.db'"; \
	docker cp $$CONTAINER:/tmp/app-backup.db $(ENV_DIR)/backups/app-$$TIMESTAMP.db; \
	docker exec $$CONTAINER rm -f /tmp/app-backup.db; \
	echo "Backup written to $(ENV_DIR)/backups/app-$$TIMESTAMP.db"

.PHONY: server-caddy-logs
server-caddy-logs:
	docker logs -f --tail=200 $(COMPOSE_PROJECT)-caddy

.PHONY: server-dev-caddy-logs server-test-caddy-logs server-prod-caddy-logs
server-dev-caddy-logs:
	$(MAKE) server-caddy-logs ENV=development

server-test-caddy-logs:
	$(MAKE) server-caddy-logs ENV=test

server-prod-caddy-logs:
	$(MAKE) server-caddy-logs ENV=production
	
.PHONY: docker-df
docker-df:
	docker system df

.PHONY: docker-prune-build-cache
docker-prune-build-cache:
	docker builder prune -f
	docker image prune -f
	docker container prune -f

server-dev-pull:
	$(MAKE) server-pull ENV=development
server-dev-build:
	$(MAKE) server-build ENV=development
server-dev-up:
	$(MAKE) server-up ENV=development
server-dev-down:
	$(MAKE) server-down ENV=development
server-dev-logs:
	$(MAKE) server-logs ENV=development
server-dev-smoke:
	$(MAKE) server-smoke ENV=development
server-dev-admin-test:
	$(MAKE) server-admin-test ENV=development
server-dev-backup:
	$(MAKE) server-backup ENV=development
server-dev-dns-check:
	$(MAKE) server-dns-check ENV=development
server-dev-cert-check:
	$(MAKE) server-cert-check ENV=development

server-test-pull:
	$(MAKE) server-pull ENV=test
server-test-build:
	$(MAKE) server-build ENV=test
server-test-up:
	$(MAKE) server-up ENV=test
server-test-down:
	$(MAKE) server-down ENV=test
server-test-logs:
	$(MAKE) server-logs ENV=test
server-test-smoke:
	$(MAKE) server-smoke ENV=test
server-test-admin-test:
	$(MAKE) server-admin-test ENV=test
server-test-backup:
	$(MAKE) server-backup ENV=test
server-test-dns-check:
	$(MAKE) server-dns-check ENV=test
server-test-cert-check:
	$(MAKE) server-cert-check ENV=test

server-prod-pull:
	$(MAKE) server-pull ENV=production
server-prod-build:
	$(MAKE) server-build ENV=production
server-prod-up:
	$(MAKE) server-up ENV=production
server-prod-down:
	$(MAKE) server-down ENV=production
server-prod-logs:
	$(MAKE) server-logs ENV=production
server-prod-smoke:
	$(MAKE) server-smoke ENV=production
server-prod-admin-test:
	$(MAKE) server-admin-test ENV=production
server-prod-backup:
	$(MAKE) server-backup ENV=production
server-prod-dns-check:
	$(MAKE) server-dns-check ENV=production
server-prod-cert-check:
	$(MAKE) server-cert-check ENV=production

.PHONY: edge-init
edge-init:
	mkdir -p $(EDGE_DIR)
	cp -R edge/* $(EDGE_DIR)/
	@echo "Initialized edge proxy folder at $(EDGE_DIR)"

.PHONY: edge-sync
edge-sync:
	mkdir -p $(EDGE_DIR)
	rsync -av --delete edge/ $(EDGE_DIR)/
	@echo "Synced edge/ to $(EDGE_DIR)"

.PHONY: edge-up
edge-up:
	cd $(EDGE_DIR) && docker compose -f docker-compose.edge.yml up -d

.PHONY: edge-down
edge-down:
	cd $(EDGE_DIR) && docker compose -f docker-compose.edge.yml down

.PHONY: edge-reload
edge-reload:
	docker exec callitcureit-edge-caddy caddy reload --config /etc/caddy/Caddyfile

.PHONY: edge-restart
edge-restart:
	cd $(EDGE_DIR) && docker compose -f docker-compose.edge.yml up -d --force-recreate

.PHONY: edge-logs
edge-logs:
	docker logs -f --tail=200 callitcureit-edge-caddy

.PHONY: edge-diff
edge-diff:
	@echo "Comparing repo edge/ with live $(EDGE_DIR)"
	diff -ru edge $(EDGE_DIR) || true

.PHONY: edge-deploy
edge-deploy: edge-sync edge-up edge-reload
	@echo "Edge proxy deployed."