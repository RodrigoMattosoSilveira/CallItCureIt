#!/usr/bin/env bash
set -euo pipefail

SERVICE="${SERVICE:-backend}"

docker compose exec "$SERVICE" sh -lc "
  set -e

  if [ -z \"\${ADMIN_EMAIL:-}\" ] || [ -z \"\${ADMIN_PASSWORD:-}\" ]; then
    echo '❌ ADMIN_EMAIL and ADMIN_PASSWORD must be set in .env'
    exit 1
  fi

  /app/create-admin
"