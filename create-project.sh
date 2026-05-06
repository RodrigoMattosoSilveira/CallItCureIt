#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="legal-objection-trainer"

echo "Creating project structure: ${ROOT_DIR}"

# Backend directories
mkdir -p "${ROOT_DIR}/backend/cmd/api"
mkdir -p "${ROOT_DIR}/backend/api/generated"
mkdir -p "${ROOT_DIR}/backend/internal/config"
mkdir -p "${ROOT_DIR}/backend/internal/db"
mkdir -p "${ROOT_DIR}/backend/internal/httpx"
mkdir -p "${ROOT_DIR}/backend/internal/scenarios"
mkdir -p "${ROOT_DIR}/backend/internal/sessions"
mkdir -p "${ROOT_DIR}/backend/internal/objections"
mkdir -p "${ROOT_DIR}/backend/internal/scoring"
mkdir -p "${ROOT_DIR}/backend/internal/feedback"
mkdir -p "${ROOT_DIR}/backend/internal/llm"
mkdir -p "${ROOT_DIR}/backend/migrations"

# Frontend directories
mkdir -p "${ROOT_DIR}/frontend/src/app"
mkdir -p "${ROOT_DIR}/frontend/src/api/generated"
mkdir -p "${ROOT_DIR}/frontend/src/features/scenarios"
mkdir -p "${ROOT_DIR}/frontend/src/features/sessions"
mkdir -p "${ROOT_DIR}/frontend/src/features/debrief"
mkdir -p "${ROOT_DIR}/frontend/src/features/admin"
mkdir -p "${ROOT_DIR}/frontend/src/components"

# Docs directory
mkdir -p "${ROOT_DIR}/docs"

# Backend files
touch "${ROOT_DIR}/backend/cmd/api/main.go"
touch "${ROOT_DIR}/backend/api/openapi.yaml"
touch "${ROOT_DIR}/backend/api/generated/api.gen.go"
touch "${ROOT_DIR}/backend/go.mod"

touch "${ROOT_DIR}/backend/migrations/000001_init_schema.up.sql"
touch "${ROOT_DIR}/backend/migrations/000001_init_schema.down.sql"
touch "${ROOT_DIR}/backend/migrations/000002_seed_reference_data.up.sql"
touch "${ROOT_DIR}/backend/migrations/000002_seed_reference_data.down.sql"

# Frontend files
touch "${ROOT_DIR}/frontend/src/app/router.tsx"
touch "${ROOT_DIR}/frontend/src/app/providers.tsx"
touch "${ROOT_DIR}/frontend/src/api/client.ts"
touch "${ROOT_DIR}/frontend/src/api/generated/schema.ts"
touch "${ROOT_DIR}/frontend/src/main.tsx"
touch "${ROOT_DIR}/frontend/package.json"

# Docs files
touch "${ROOT_DIR}/docs/architecture.md"
touch "${ROOT_DIR}/docs/api.md"
touch "${ROOT_DIR}/docs/objection-taxonomy.md"

echo "Done."
echo
echo "Created:"
find "${ROOT_DIR}" | sort