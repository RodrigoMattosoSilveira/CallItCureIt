#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"

cd "${FRONTEND_DIR}"

if [[ ! -f ".env" ]]; then
  cat > .env <<'EOF'
VITE_API_BASE_URL=/api/v1
EOF
fi

npm run dev