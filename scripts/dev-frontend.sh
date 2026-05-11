#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}/frontend"

export VITE_API_BASE_URL="${VITE_API_BASE_URL:-http://localhost:8080/api/v1}"

echo "Starting frontend..."
echo "VITE_API_BASE_URL=$VITE_API_BASE_URL"
echo

npm run dev