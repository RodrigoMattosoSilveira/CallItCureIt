#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}/backend"

export DATABASE_PATH="${DATABASE_PATH:-data/app.db}"
export PORT="${PORT:-8080}"
export LLM_COACHING_ENABLED="${LLM_COACHING_ENABLED:-false}"

echo "Starting backend..."
echo "DATABASE_PATH=$DATABASE_PATH"
echo "PORT=$PORT"
echo "LLM_COACHING_ENABLED=$LLM_COACHING_ENABLED"
echo

go run ./cmd/api