#!/usr/bin/env bash
set -euo pipefail

# Creates or updates local development env files:
#   backend/.env
#   frontend/.env

BACKEND_ENV_FILE="${BACKEND_ENV_FILE:-backend/.env}"
FRONTEND_ENV_FILE="${FRONTEND_ENV_FILE:-frontend/.env}"

mkdir -p "$(dirname "$BACKEND_ENV_FILE")"
mkdir -p "$(dirname "$FRONTEND_ENV_FILE")"

touch "$BACKEND_ENV_FILE"
touch "$FRONTEND_ENV_FILE"

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

# Backend local development
set_or_update_env "$BACKEND_ENV_FILE" "APP_ENV" "local"
set_or_update_env "$BACKEND_ENV_FILE" "PORT" "8080"
set_or_update_env "$BACKEND_ENV_FILE" "DATABASE_PATH" "data/app.db"

set_or_update_env "$BACKEND_ENV_FILE" "JWT_SECRET" "dev-secret-change-me"
set_or_update_env "$BACKEND_ENV_FILE" "JWT_ISSUER" "call-it-cure-it"
set_or_update_env "$BACKEND_ENV_FILE" "JWT_EXPIRATION_MINUTES" "480"

set_or_update_env "$BACKEND_ENV_FILE" "DEV_SEED_ADMIN" "true"
set_or_update_env "$BACKEND_ENV_FILE" "DEV_ADMIN_EMAIL" "admin@example.com"
set_or_update_env "$BACKEND_ENV_FILE" "DEV_ADMIN_PASSWORD" "admin123"
set_or_update_env "$BACKEND_ENV_FILE" "DEV_ADMIN_NAME" "Admin User"

set_or_update_env "$BACKEND_ENV_FILE" "LLM_COACHING_ENABLED" "false"
set_or_update_env "$BACKEND_ENV_FILE" "OPENAI_API_KEY" ""
set_or_update_env "$BACKEND_ENV_FILE" "OPENAI_MODEL" "gpt-5.1-mini"
set_or_update_env "$BACKEND_ENV_FILE" "OPENAI_BASE_URL" "https://api.openai.com/v1"
set_or_update_env "$BACKEND_ENV_FILE" "OPENAI_TIMEOUT_SECONDS" "20"

set_or_update_env "$BACKEND_ENV_FILE" "CORS_ALLOW_ORIGINS" "http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173"

# Frontend local development
set_or_update_env "$FRONTEND_ENV_FILE" "VITE_API_BASE_URL" "/api/v1"

echo "Created/updated:"
echo "  ${BACKEND_ENV_FILE}"
echo "  ${FRONTEND_ENV_FILE}"
echo
echo "Backend dev admin:"
echo "  DEV_ADMIN_EMAIL=admin@example.com"
echo "  DEV_ADMIN_PASSWORD=admin123"
echo
echo "Frontend API:"
echo "  VITE_API_BASE_URL=/api/v1"