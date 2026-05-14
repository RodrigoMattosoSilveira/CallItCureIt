#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env.production}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing ${ENV_FILE}. Run ./scripts/init-env.sh first."
  exit 1
fi

docker compose --env-file "$ENV_FILE" -f docker-compose.prod.yml build