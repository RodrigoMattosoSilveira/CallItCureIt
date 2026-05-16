# Call It Cure It Environment Checklist

This is the consolidated environment, deployment, and operations checklist for **Call It Cure It**.

It integrates local development, GitHub Issue/branch/PR workflow, three server environments, one shared Caddy reverse proxy, Docker Compose deployment, environment-specific folders, branches, env files, container names, SQLite volumes, and Makefile-first operations.

## Environment Map

| Environment | DNS | Git branch | Server folder | Compose project | Container prefix |
|---|---|---|---|---|---|
| Local | localhost / LAN IP | feature branch | developer machine | n/a | n/a |
| Development | `dev.callitcureit.com` | `development` | `/opt/CallItCureIt/development` | `callitcureit-dev` | `callitcureit-dev` |
| Test | `tst.callitcureit.com` | `test` | `/opt/CallItCureIt/test` | `callitcureit-tst` | `callitcureit-tst` |
| Production | `app.callitcureit.com` | `production` | `/opt/CallItCureIt/production` | `callitcureit-prd` | `callitcureit-prd` |

The shared reverse proxy lives at:

```text
/opt/CallItCureIt/reverse-proxy
```

and owns public ports `80` and `443`. The app stacks do **not** expose public ports. They join the shared Docker network and are reached by Caddy internally.

---

## 1. Core Rules

Use `make` for routine operations. Direct shell commands are acceptable for tasks outside the app automation boundary, such as generating SSH keys, adding GitHub deploy keys, editing DNS records, installing Docker on a new server, and editing secret values.

Do **not** use `backend/cmd/create-admin`. Admin bootstrap is handled by the API at startup:

```text
config.Load()
  -> authService.EnsureDevAdmin(...)
  -> create admin if DEV_SEED_ADMIN=true and user is missing
```

Frontend API base URL should be:

```env
VITE_API_BASE_URL=/api/v1
```

Do not use `localhost` or LAN IPs in production frontend builds.

Only one Caddy container should bind ports `80` and `443`. Do not run one Caddy container per environment.

---

## 2. GitHub Issue, Branch, PR, and Promotion Workflow

Normal software change workflow:

```text
1. Write a GitHub Issue.
2. Create a feature branch, usually from development.
3. Check out that feature branch locally.
4. Implement and test locally.
5. Commit and push the feature branch.
6. Open a pull request back into the source branch, usually development.
7. Merge after review.
8. Deploy development branch to dev.callitcureit.com.
```

Promotion workflow:

```text
Feature branch
  -> PR into development
  -> deploy development to dev.callitcureit.com
  -> validate
  -> PR/merge development into test
  -> deploy test to tst.callitcureit.com
  -> users and engineers validate
  -> PR/merge test into production
  -> deploy production to app.callitcureit.com
```

Emergency hotfix branches directly from `test` or `production` are out of scope for now.

---

## 3. Expected Repository Files

```text
backend/
  Dockerfile
  docker-entrypoint.sh
  .dockerignore
  .env.example
  .env.production.example
  migrations/
  cmd/api/main.go

frontend/
  Dockerfile
  nginx.conf
  .dockerignore
  .env.example
  .env.production.example
  vite.config.ts

reverse-proxy/
  docker-compose.proxy.yml
  Caddyfile

deploy/
  Caddyfile.example

scripts/
  init-dev-env.sh
  init-server-env.sh
  dev-backend.sh
  dev-frontend.sh

Makefile
docker-compose.server.yml
```

`backend/cmd/create-admin` and `backend/cmd/create-admin.disabled` should not exist.

---

## 4. Local Development Environment

Purpose: individual software engineers implement work for GitHub Issues.

Setup:

```bash
make doctor
make local-init-env
make local-db-init
make local-check
```

Run services in two terminals:

```bash
make local-backend
make local-frontend
```

Open:

```text
http://localhost:5173
```

For iPhone/LAN testing, find the Mac LAN IP manually:

```bash
ipconfig getifaddr en0
```

Then:

```bash
make local-lan-smoke LAN_HOST=192.168.2.154
```

Open on iPhone:

```text
http://192.168.2.154:5173/login
```

Do not use `localhost` on the iPhone.

---

## 5. Server Layout

```text
/opt/CallItCureIt/
  reverse-proxy/
  development/
  test/
  production/
```

Each app folder is a separate clone:

```text
/opt/CallItCureIt/development  -> branch development
/opt/CallItCureIt/test         -> branch test
/opt/CallItCureIt/production   -> branch production
```

---

## 6. Shared Docker Network

All app stacks and Caddy must join:

```text
callitcureit-proxy
```

Create it with:

```bash
make proxy-network
```

---

## 7. Shared Reverse Proxy

Location:

```text
/opt/CallItCureIt/reverse-proxy
```

Operations:

```bash
make proxy-up
make proxy-reload
make proxy-logs
```

---

# 8. Required Files

## 8.1 `backend/Dockerfile`

```dockerfile
# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/call-it-cure-it-api ./cmd/api

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates sqlite3 curl \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/call-it-cure-it-api /app/call-it-cure-it-api
COPY migrations /app/migrations
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh \
    && mkdir -p /app/data
ENV PORT=8080
ENV DATABASE_PATH=/app/data/app.db
ENV LLM_COACHING_ENABLED=false
EXPOSE 8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
```

## 8.2 `backend/docker-entrypoint.sh`

```sh
#!/usr/bin/env sh
set -eu

DB_PATH="${DATABASE_PATH:-/app/data/app.db}"
mkdir -p "$(dirname "$DB_PATH")"

echo "Applying database migrations to ${DB_PATH}..."
for migration in /app/migrations/*.up.sql; do
  echo "Applying ${migration}"
  sqlite3 "$DB_PATH" < "$migration"
done

echo "Starting API..."
exec /app/call-it-cure-it-api
```

Make executable:

```bash
chmod +x backend/docker-entrypoint.sh
```

This simple entrypoint assumes migrations are safe to rerun. If a migration is not idempotent, add a `schema_migrations` table later.

## 8.3 `backend/.dockerignore`

```gitignore
data/
tmp/
.env
.env.*
*.db
*.db-shm
*.db-wal
coverage.out
```

## 8.4 `frontend/Dockerfile`

```dockerfile
# syntax=docker/dockerfile:1

FROM node:22-bookworm-slim AS build
WORKDIR /src
COPY package*.json ./
RUN npm ci
COPY . .
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN npm run build

FROM nginx:1.27-alpine AS runtime
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 8.5 `frontend/nginx.conf`

```nginx
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    client_max_body_size 10m;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /healthz {
        return 200 "ok\n";
        add_header Content-Type text/plain;
    }
}
```

## 8.6 `frontend/.dockerignore`

```gitignore
node_modules/
dist/
.env
.env.*
coverage/
```

## 8.7 `docker-compose.server.yml`

This is the app stack template used by development, test, and production. It does not include Caddy.

```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: ${CONTAINER_PREFIX}-backend
    restart: unless-stopped
    environment:
      PORT: "8080"
      DATABASE_PATH: "/app/data/app.db"
      JWT_SECRET: "${JWT_SECRET}"
      JWT_ISSUER: "${JWT_ISSUER}"
      JWT_EXPIRATION_MINUTES: "${JWT_EXPIRATION_MINUTES:-480}"
      DEV_SEED_ADMIN: "${DEV_SEED_ADMIN:-false}"
      DEV_ADMIN_EMAIL: "${DEV_ADMIN_EMAIL:-}"
      DEV_ADMIN_PASSWORD: "${DEV_ADMIN_PASSWORD:-}"
      DEV_ADMIN_NAME: "${DEV_ADMIN_NAME:-}"
      LLM_COACHING_ENABLED: "${LLM_COACHING_ENABLED:-false}"
      OPENAI_API_KEY: "${OPENAI_API_KEY:-}"
      OPENAI_MODEL: "${OPENAI_MODEL:-gpt-5.1-mini}"
      OPENAI_BASE_URL: "${OPENAI_BASE_URL:-https://api.openai.com/v1}"
      OPENAI_TIMEOUT_SECONDS: "${OPENAI_TIMEOUT_SECONDS:-20}"
      CORS_ALLOW_ORIGINS: "${CORS_ALLOW_ORIGINS}"
    volumes:
      - backend-data:/app/data
    expose:
      - "8080"
    networks:
      - app-net
      - callitcureit-proxy
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://localhost:8080/api/v1/healthz"]
      interval: 20s
      timeout: 5s
      retries: 5

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        VITE_API_BASE_URL: "/api/v1"
    container_name: ${CONTAINER_PREFIX}-frontend
    restart: unless-stopped
    expose:
      - "80"
    depends_on:
      backend:
        condition: service_healthy
    networks:
      - app-net
      - callitcureit-proxy

volumes:
  backend-data:

networks:
  app-net:
  callitcureit-proxy:
    external: true
```

## 8.8 `reverse-proxy/docker-compose.proxy.yml`

```yaml
services:
  caddy:
    image: caddy:2.8-alpine
    container_name: callitcureit-caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy-data:/data
      - caddy-config:/config
    networks:
      - callitcureit-proxy

volumes:
  caddy-data:
  caddy-config:

networks:
  callitcureit-proxy:
    external: true
```

## 8.9 `reverse-proxy/Caddyfile`

```caddyfile
dev.callitcureit.com {
    encode gzip zstd
    handle_path /api/* {
        reverse_proxy callitcureit-dev-backend:8080
    }
    handle {
        reverse_proxy callitcureit-dev-frontend:80
    }
}

tst.callitcureit.com {
    encode gzip zstd
    handle_path /api/* {
        reverse_proxy callitcureit-tst-backend:8080
    }
    handle {
        reverse_proxy callitcureit-tst-frontend:80
    }
}

app.callitcureit.com {
    encode gzip zstd
    handle_path /api/* {
        reverse_proxy callitcureit-prd-backend:8080
    }
    handle {
        reverse_proxy callitcureit-prd-frontend:80
    }
}
```

---

# 9. Required Scripts

## 9.1 `scripts/init-dev-env.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

BACKEND_ENV="backend/.env"
FRONTEND_ENV="frontend/.env"
mkdir -p backend frontend
touch "$BACKEND_ENV" "$FRONTEND_ENV"

set_or_update_env() {
  local file="$1"
  local key="$2"
  local value="$3"
  local escaped_value
  escaped_value="$(printf '%s' "$value" | sed 's/[\/&]/\\&/g')"
  if grep -qE "^[[:space:]]*${key}=" "$file"; then
    sed -i.bak -E "s|^[[:space:]]*${key}=.*|${key}=${escaped_value}|" "$file"
  else
    printf '%s=%s\n' "$key" "$value" >> "$file"
  fi
  rm -f "${file}.bak"
}

set_or_update_env "$BACKEND_ENV" "APP_ENV" "local"
set_or_update_env "$BACKEND_ENV" "PORT" "8080"
set_or_update_env "$BACKEND_ENV" "DATABASE_PATH" "data/app.db"
set_or_update_env "$BACKEND_ENV" "JWT_SECRET" "dev-secret-change-me"
set_or_update_env "$BACKEND_ENV" "JWT_ISSUER" "call-it-cure-it"
set_or_update_env "$BACKEND_ENV" "JWT_EXPIRATION_MINUTES" "480"
set_or_update_env "$BACKEND_ENV" "DEV_SEED_ADMIN" "true"
set_or_update_env "$BACKEND_ENV" "DEV_ADMIN_EMAIL" "admin@example.com"
set_or_update_env "$BACKEND_ENV" "DEV_ADMIN_PASSWORD" "admin123"
set_or_update_env "$BACKEND_ENV" "DEV_ADMIN_NAME" "Admin User"
set_or_update_env "$BACKEND_ENV" "LLM_COACHING_ENABLED" "false"
set_or_update_env "$BACKEND_ENV" "OPENAI_API_KEY" ""
set_or_update_env "$BACKEND_ENV" "OPENAI_MODEL" "gpt-5.1-mini"
set_or_update_env "$BACKEND_ENV" "OPENAI_BASE_URL" "https://api.openai.com/v1"
set_or_update_env "$BACKEND_ENV" "OPENAI_TIMEOUT_SECONDS" "20"
set_or_update_env "$BACKEND_ENV" "CORS_ALLOW_ORIGINS" "http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173"
set_or_update_env "$FRONTEND_ENV" "VITE_API_BASE_URL" "/api/v1"

echo "Initialized local environment files."
```

## 9.2 `scripts/init-server-env.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-}"

if [[ -z "$ENV_NAME" ]]; then
  echo "Usage: $0 development|test|production"
  exit 1
fi

case "$ENV_NAME" in
  development)
    ENV_FILE=".env.development"
    APP_ENV="development"
    APP_DOMAIN="dev.callitcureit.com"
    CONTAINER_PREFIX="callitcureit-dev"
    JWT_ISSUER="call-it-cure-it-development"
    DEV_ADMIN_EMAIL="admin-dev@callitcureit.com"
    DEV_ADMIN_NAME="Development Admin"
    ;;
  test)
    ENV_FILE=".env.test"
    APP_ENV="test"
    APP_DOMAIN="tst.callitcureit.com"
    CONTAINER_PREFIX="callitcureit-tst"
    JWT_ISSUER="call-it-cure-it-test"
    DEV_ADMIN_EMAIL="admin-tst@callitcureit.com"
    DEV_ADMIN_NAME="Test Admin"
    ;;
  production)
    ENV_FILE=".env.production"
    APP_ENV="production"
    APP_DOMAIN="app.callitcureit.com"
    CONTAINER_PREFIX="callitcureit-prd"
    JWT_ISSUER="call-it-cure-it"
    DEV_ADMIN_EMAIL="admin@callitcureit.com"
    DEV_ADMIN_NAME="Production Admin"
    ;;
  *)
    echo "Invalid environment: $ENV_NAME"
    exit 1
    ;;
esac

touch "$ENV_FILE"

set_or_update_env() {
  local key="$1"
  local value="$2"
  local escaped_value
  escaped_value="$(printf '%s' "$value" | sed 's/[\/&]/\\&/g')"
  if grep -qE "^[[:space:]]*${key}=" "$ENV_FILE"; then
    sed -i.bak -E "s|^[[:space:]]*${key}=.*|${key}=${escaped_value}|" "$ENV_FILE"
  else
    printf '%s=%s\n' "$key" "$value" >> "$ENV_FILE"
  fi
}

set_or_update_env "APP_ENV" "$APP_ENV"
set_or_update_env "APP_DOMAIN" "$APP_DOMAIN"
set_or_update_env "CONTAINER_PREFIX" "$CONTAINER_PREFIX"
set_or_update_env "PORT" "8080"
set_or_update_env "DATABASE_PATH" "/app/data/app.db"

if ! grep -qE "^[[:space:]]*JWT_SECRET=." "$ENV_FILE"; then
  set_or_update_env "JWT_SECRET" "$(openssl rand -base64 48)"
fi

set_or_update_env "JWT_ISSUER" "$JWT_ISSUER"
set_or_update_env "JWT_EXPIRATION_MINUTES" "480"
set_or_update_env "DEV_SEED_ADMIN" "true"
set_or_update_env "DEV_ADMIN_EMAIL" "$DEV_ADMIN_EMAIL"
set_or_update_env "DEV_ADMIN_PASSWORD" "change-this-password-immediately"
set_or_update_env "DEV_ADMIN_NAME" "$DEV_ADMIN_NAME"
set_or_update_env "LLM_COACHING_ENABLED" "false"
set_or_update_env "OPENAI_API_KEY" ""
set_or_update_env "OPENAI_MODEL" "gpt-5.1-mini"
set_or_update_env "OPENAI_BASE_URL" "https://api.openai.com/v1"
set_or_update_env "OPENAI_TIMEOUT_SECONDS" "20"
set_or_update_env "CORS_ALLOW_ORIGINS" "https://${APP_DOMAIN}"

rm -f "${ENV_FILE}.bak"

echo "Created/updated ${ENV_FILE}"
echo "IMPORTANT: edit ${ENV_FILE} and set a strong DEV_ADMIN_PASSWORD."
echo "For production, after first successful admin login, set DEV_SEED_ADMIN=false."
```

## 9.3 `scripts/dev-backend.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND_DIR="${PROJECT_ROOT}/backend"
cd "${BACKEND_DIR}"

export PORT="${PORT:-8080}"
export DATABASE_PATH="${DATABASE_PATH:-data/app.db}"
export DEV_SEED_ADMIN="${DEV_SEED_ADMIN:-true}"
export DEV_ADMIN_EMAIL="${DEV_ADMIN_EMAIL:-admin@example.com}"
export DEV_ADMIN_PASSWORD="${DEV_ADMIN_PASSWORD:-admin123}"
export DEV_ADMIN_NAME="${DEV_ADMIN_NAME:-Admin User}"
export JWT_SECRET="${JWT_SECRET:-dev-secret-change-me}"
export JWT_ISSUER="${JWT_ISSUER:-call-it-cure-it}"
export JWT_EXPIRATION_MINUTES="${JWT_EXPIRATION_MINUTES:-480}"
export LLM_COACHING_ENABLED="${LLM_COACHING_ENABLED:-false}"
export OPENAI_API_KEY="${OPENAI_API_KEY:-}"
export OPENAI_MODEL="${OPENAI_MODEL:-gpt-5.1-mini}"
export OPENAI_BASE_URL="${OPENAI_BASE_URL:-https://api.openai.com/v1}"
export OPENAI_TIMEOUT_SECONDS="${OPENAI_TIMEOUT_SECONDS:-20}"
export CORS_ALLOW_ORIGINS="${CORS_ALLOW_ORIGINS:-http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173}"

mkdir -p data

go run ./cmd/api
```

## 9.4 `scripts/dev-frontend.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"
cd "${FRONTEND_DIR}"

if [[ ! -f ".env" ]]; then
  cat > .env <<'ENVEOF'
VITE_API_BASE_URL=/api/v1
ENVEOF
fi

npm run dev
```

---

# 10. Root Makefile

```make
SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

SERVER_ROOT ?= /opt/CallItCureIt
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
	@echo "Proxy: make proxy-network proxy-up proxy-logs proxy-reload"
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

.PHONY: proxy-network
proxy-network:
	docker network inspect callitcureit-proxy >/dev/null 2>&1 || docker network create callitcureit-proxy

.PHONY: proxy-up
proxy-up: proxy-network
	cd $(PROXY_DIR) && docker compose -f docker-compose.proxy.yml up -d

.PHONY: proxy-down
proxy-down:
	cd $(PROXY_DIR) && docker compose -f docker-compose.proxy.yml down

.PHONY: proxy-logs
proxy-logs:
	docker logs -f --tail=200 callitcureit-caddy

.PHONY: proxy-reload
proxy-reload:
	docker exec callitcureit-caddy caddy reload --config /etc/caddy/Caddyfile

.PHONY: server-init-env
server-init-env:
	chmod +x scripts/init-server-env.sh
	./scripts/init-server-env.sh $(ENV)

.PHONY: server-pull
server-pull:
	cd $(ENV_DIR) && git checkout $(BRANCH) && git pull

.PHONY: server-build
server-build:
	cd $(ENV_DIR) && BUILDX_NO_DEFAULT_ATTESTATIONS=1 docker compose -p $(COMPOSE_PROJECT) --env-file $(ENV_FILE) -f docker-compose.server.yml build --progress=plain

.PHONY: server-up
server-up: proxy-network
	cd $(ENV_DIR) && docker compose -p $(COMPOSE_PROJECT) --env-file $(ENV_FILE) -f docker-compose.server.yml up -d

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
```

---

# 11. Server Bootstrap Procedure

Manual prerequisites before the repository is available:

```bash
sudo apt update
sudo apt upgrade -y
sudo apt install -y ca-certificates curl git ufw fail2ban sqlite3 jq openssl
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker deploy
```

Reconnect, then configure firewall:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Generate GitHub deploy key manually:

```bash
ssh-keygen -t ed25519 -C "hetzner-callitcureit-deploy" -f ~/.ssh/callitcureit_deploy
```

Add the public key to GitHub repository deploy keys.

Clone app folders:

```bash
mkdir -p /opt/CallItCureIt
cd /opt/CallItCureIt

git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git development
git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git test
git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git production

cd /opt/CallItCureIt/development && git checkout development
cd /opt/CallItCureIt/test && git checkout test
cd /opt/CallItCureIt/production && git checkout production
```

Create shared reverse proxy folder:

```bash
mkdir -p /opt/CallItCureIt/reverse-proxy
cp /opt/CallItCureIt/production/reverse-proxy/docker-compose.proxy.yml /opt/CallItCureIt/reverse-proxy/
cp /opt/CallItCureIt/production/reverse-proxy/Caddyfile /opt/CallItCureIt/reverse-proxy/
```

---

# 12. Server Environment Initialization

Development:

```bash
cd /opt/CallItCureIt/development
make server-init-env ENV=development
nano .env.development
make server-dev-build
make server-dev-up
```

Test:

```bash
cd /opt/CallItCureIt/test
make server-init-env ENV=test
nano .env.test
make server-test-build
make server-test-up
```

Production:

```bash
cd /opt/CallItCureIt/production
make server-init-env ENV=production
nano .env.production
make server-prod-build
make server-prod-up
```

Shared proxy:

```bash
cd /opt/CallItCureIt/production
make proxy-network
cd /opt/CallItCureIt/reverse-proxy
docker compose -f docker-compose.proxy.yml up -d
```

---

# 13. DNS Checklist

All three A records must point to the Hetzner server:

```text
dev.callitcureit.com  A  5.78.208.230
tst.callitcureit.com  A  5.78.208.230
app.callitcureit.com  A  5.78.208.230
```

Check:

```bash
make server-dev-dns-check
make server-test-dns-check
make server-prod-dns-check
```

Also ensure no incorrect AAAA records point elsewhere.

---

# 14. Validation Checklist

Inspect network:

```bash
docker network inspect callitcureit-proxy
```

Expected containers include:

```text
callitcureit-caddy
callitcureit-dev-backend
callitcureit-dev-frontend
callitcureit-tst-backend
callitcureit-tst-frontend
callitcureit-prd-backend
callitcureit-prd-frontend
```

Validate development:

```bash
make server-dev-smoke
make server-dev-admin-test
make server-dev-cert-check
```

Validate test:

```bash
make server-test-smoke
make server-test-admin-test
make server-test-cert-check
```

Validate production:

```bash
make server-prod-smoke
make server-prod-admin-test
make server-prod-cert-check
```

After first successful production admin login, edit `.env.production`:

```env
DEV_SEED_ADMIN=false
```

Then restart:

```bash
make server-prod-up
```

---

# 15. Deployment and Promotion Procedures

Deploy development:

```bash
cd /opt/CallItCureIt/development
make server-dev-pull
make server-dev-build
make server-dev-up
make server-dev-smoke
```

Promote to test after merging `development` into `test`:

```bash
cd /opt/CallItCureIt/test
make server-test-pull
make server-test-build
make server-test-up
make server-test-smoke
```

Promote to production after merging `test` into `production`:

```bash
cd /opt/CallItCureIt/production
make server-prod-pull
make server-prod-build
make server-prod-up
make server-prod-smoke
make server-prod-admin-test
```

---

# 16. Backups

Each environment has its own SQLite volume.

```bash
make server-dev-backup
make server-test-backup
make server-prod-backup
```

Backups are written to:

```text
/opt/CallItCureIt/development/backups/
/opt/CallItCureIt/test/backups/
/opt/CallItCureIt/production/backups/
```

Copy production backups off-server regularly.

---

# 17. Troubleshooting

Caddy cannot route to an environment:

```bash
docker network inspect callitcureit-proxy
```

TLS issue:

```bash
make server-prod-cert-check
make proxy-logs
make server-prod-dns-check
```

Backend says `no such table: users`:

```bash
make server-prod-build
make server-prod-up
```

Docker build stuck on provenance: use Makefile build targets; they set `BUILDX_NO_DEFAULT_ATTESTATIONS=1`.

Docker commands require sudo:

```bash
sudo usermod -aG docker deploy
```

Then log out and back in.

---

# 18. Final Environment Checklist

## Local

```text
[ ] make doctor passes
[ ] make local-init-env
[ ] make local-db-init
[ ] make local-check
[ ] make local-backend
[ ] make local-frontend
[ ] make local-smoke
[ ] make local-admin-test
[ ] Browser training flow works
[ ] Browser admin flow works
```

## Development Server

```text
[ ] dev.callitcureit.com DNS points to server
[ ] /opt/CallItCureIt/development exists
[ ] development branch checked out
[ ] .env.development exists
[ ] make server-dev-build
[ ] make server-dev-up
[ ] make server-dev-smoke
[ ] make server-dev-admin-test
[ ] Engineers validate dev.callitcureit.com
```
### Verify dev is healthy

**Run:**

```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

**Expected:**
```bash
callitcureit-dev-backend    Up ... healthy
callitcureit-dev-frontend   Up ...
callitcureit-caddy          Up ...
```
**Then test backend inside the container:**
```bash
docker exec -it callitcureit-dev-backend curl -i http://localhost:8080/api/v1/healthz
```

**Then test through Caddy:**
```bash
curl -i https://dev.callitcureit.com/api/v1/healthz
```
**Then rerun:**
```bash
make server-dev-smoke
make server-dev-admin-test
```

## Test Server

```text
[ ] tst.callitcureit.com DNS points to server
[ ] /opt/CallItCureIt/test exists
[ ] test branch checked out
[ ] .env.test exists
[ ] make server-test-build
[ ] make server-test-up
[ ] make server-test-smoke
[ ] make server-test-admin-test
[ ] Users and engineers validate tst.callitcureit.com
```

## Production Server

```text
[ ] app.callitcureit.com DNS points to server
[ ] /opt/CallItCureIt/production exists
[ ] production branch checked out
[ ] .env.production exists
[ ] production secrets are strong
[ ] make server-prod-build
[ ] make server-prod-up
[ ] make server-prod-smoke
[ ] make server-prod-admin-test
[ ] DEV_SEED_ADMIN=false after first login
[ ] make server-prod-backup
[ ] Production users can use app.callitcureit.com
```

### Important caution for test and production

Do not blindly wipe test or production volumes.

**For development, this is okay:**
```bash
down -v
```

**For test/production, first make a backup:**

```bash
make server-test-backup
make server-prod-backup
```

If test or production already has an existing database created before schema_migrations, we should either:
1. manually insert already-applied migration filenames into schema_migrations, or
2. add a one-time bootstrap command that marks existing migrations as applied.

For dev, reset is simplest. **For production, preserve the database**.

## Shared Proxy

```text
[ ] /opt/CallItCureIt/reverse-proxy exists
[ ] callitcureit-proxy network exists
[ ] callitcureit-caddy is running
[ ] Caddyfile has dev/tst/app domains
[ ] all app containers are attached to callitcureit-proxy
[ ] TLS certs are valid for all three domains
```

---

# 19. Summary

Final deployment model:

```text
Local machine:
  feature branches
  local testing
  PR into development

Hetzner server:
  /opt/CallItCureIt/reverse-proxy  -> one Caddy for all domains
  /opt/CallItCureIt/development    -> development branch -> dev.callitcureit.com
  /opt/CallItCureIt/test           -> test branch        -> tst.callitcureit.com
  /opt/CallItCureIt/production     -> production branch  -> app.callitcureit.com
```

Each environment has its own branch, folder, env file, Compose project, container prefix, SQLite volume, admin bootstrap values, smoke tests, and backup command.

One shared Caddy proxy routes all public traffic to the correct internal app stack.

# Debuging
## Caddy Failure
It fails somewhere in this pipe
`Browser/Caddy URL -> Caddy -> dev backend container -> Fiber route`

### 1. Confirm the dev backend itself is healthy

On the server:
```bash
docker exec -it callitcureit-dev-backend curl -i http://localhost:8080/api/v1/healthz
```
Expected:

`HTTP/1.1 200 OK`

If this returns 404, the problem is inside the dev backend container, not Caddy.

If this returns 200, continue.
### 2. Confirm Caddy can reach the dev backend container

Run:
```bash
docker exec -it callitcureit-caddy wget -S -O- http://callitcureit-dev-backend:8080/api/v1/healthz 2>&1 | head -40
```
Expected:

`HTTP/1.1 200 OK`

If this fails, the problem is the shared Docker network or container name.

Check the shared network:
```bash
docker network inspect callitcureit-proxy \
  --format '{{range .Containers}}{{.Name}}{{"\n"}}{{end}}'
```
Expected to include:
```
callitcureit-caddy
callitcureit-dev-backend
callitcureit-dev-frontend
```
### 3. Confirm Caddy is using the expected config inside the container

Run:

docker exec callitcureit-caddy cat /etc/caddy/Caddyfile

The dev block must be:

dev.callitcureit.com {
    encode gzip zstd

    handle /api/* {
        reverse_proxy callitcureit-dev-backend:8080
    }

    handle {
        reverse_proxy callitcureit-dev-frontend:80
    }
}

If the Caddyfile inside the container is different from /opt/CallItCureIt/reverse-proxy/Caddyfile, then the proxy container is mounted from the wrong folder or was not recreated.

Reload Caddy:
```bash
docker exec callitcureit-caddy caddy reload --config /etc/caddy/Caddyfile
```
Then retest:
```bash
curl -i https://dev.callitcureit.com/api/v1/healthz
```
### 4. Test inside the dev backend container

Run:
```bash
docker exec -it callitcureit-dev-backend curl -i http://localhost:8080/api/v1/healthz
```
**Then also run:**
```bash
docker exec -it callitcureit-dev-backend curl -i http://localhost:8080/healthz
```
**If this works:**

`http://localhost:8080/healthz`

but this fails:

`http://localhost:8080/api/v1/healthz`

then your smoke test and backend route are out of sync.
### 5. Fix cmd/api/main.go

Open the backend file on the development branch:
```bash
cd /opt/CallItCureIt/development
grep -R "healthz" -n backend/cmd backend/internal
```
You probably have something like:
```go
app.Get("/healthz", func(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
})
```

Change this route:
```go
healthHandler := func(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

app.Get("/healthz", healthHandler)
app.Get("/api/v1/healthz", healthHandler)
```