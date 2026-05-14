# Introduction
This is a deterministic architecture: one shared file, one environment-specific override file, and one render script that generates the exact .env files used by Docker/backend/frontend.

It gives us this structure:

```
env/
  common.env
  dev.env
  tst.env
  prd.env

scripts/
  render-env.sh
  render-all-envs.sh
  print-env.sh
```
The rule is simple:

```
# Environment-specific values override common values.
final env = env/common.env + env/<environment>.env
```

# Create Common Parameters
`env/common.env`
```bash
# Common values shared by all environments

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_ISSUER=call-it-cure-it
JWT_EXPIRATION_MINUTES=480

OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

# Frontend should call the backend through the same origin in Docker/prod.
VITE_API_BASE_URL=/api/v1
```

# Create Development Overrides
`env/dev.env`
```bash
# Development overrides
APP_ENV=dev
APP_DOMAIN=localhost

DATABASE_PATH=data/app.db

JWT_SECRET=dev-secret-change-me

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@example.com
DEV_ADMIN_PASSWORD=admin123
DEV_ADMIN_NAME=Admin User

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=

CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173
```

# Create Test/Staging Overrides
`env/tst.env`
```bash
# Test/Staging overrides
APP_ENV=tst
APP_DOMAIN=tst.yourdomain.com

JWT_SECRET=replace-with-tst-secret

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin-tst@example.com
DEV_ADMIN_PASSWORD=replace-with-tst-admin-password
DEV_ADMIN_NAME=Test Admin

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=

CORS_ALLOW_ORIGINS=https://tst.yourdomain.com
```

# Create Production Overrides
`env/prd.env`
```bash
# Production overrides
APP_ENV=prd
APP_DOMAIN=app.yourdomain.com

JWT_SECRET=replace-with-long-production-secret

# Recommended after first bootstrap:
# DEV_SEED_ADMIN=false
DEV_SEED_ADMIN=false
DEV_ADMIN_EMAIL=
DEV_ADMIN_PASSWORD=
DEV_ADMIN_NAME=

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=

CORS_ALLOW_ORIGINS=https://app.yourdomain.com
```

# Render Desired Environment
This script merges:

`env/common.env + env/<env>.env`

and writes a concrete output file.`

`scripts/render-env.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./scripts/render-env.sh dev
#   ./scripts/render-env.sh tst
#   ./scripts/render-env.sh prd
#   ./scripts/render-env.sh dev backend/.env
#
# Default output:
#   dev -> .env.dev
#   tst -> .env.tst
#   prd -> .env.prd

ENV_NAME="${1:-}"
OUTPUT_FILE="${2:-}"

if [[ -z "${ENV_NAME}" ]]; then
  echo "Usage: $0 <dev|tst|prd> [output-file]"
  exit 1
fi

COMMON_FILE="env/common.env"
ENV_FILE="env/${ENV_NAME}.env"

if [[ ! -f "${COMMON_FILE}" ]]; then
  echo "Error: missing ${COMMON_FILE}"
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Error: missing ${ENV_FILE}"
  exit 1
fi

if [[ -z "${OUTPUT_FILE}" ]]; then
  OUTPUT_FILE=".env.${ENV_NAME}"
fi

TMP_FILE="$(mktemp)"
trap 'rm -f "${TMP_FILE}"' EXIT

# Merge dotenv files.
# Later files override earlier files.
awk '
  function trim(s) {
    gsub(/^[ \t]+|[ \t]+$/, "", s)
    return s
  }

  /^[ \t]*$/ { next }
  /^[ \t]*#/ { next }

  {
    line=$0
    eq=index(line, "=")
    if (eq == 0) {
      next
    }

    key=trim(substr(line, 1, eq - 1))
    value=substr(line, eq + 1)

    if (key == "") {
      next
    }

    order[++n]=key
    values[key]=value
    seen[key]=1
  }

  END {
    print "# Generated file. Do not edit directly."
    print "# Source files: env/common.env + " ENV_FILE
    print ""

    for (i = 1; i <= n; i++) {
      key=order[i]
      if (printed[key] == 1) {
        continue
      }

      # Print keys in final value order. If repeated, value is latest.
      print key "=" values[key]
      printed[key]=1
    }
  }
' ENV_FILE="${ENV_FILE}" "${COMMON_FILE}" "${ENV_FILE}" > "${TMP_FILE}"

mkdir -p "$(dirname "${OUTPUT_FILE}")"
cp "${TMP_FILE}" "${OUTPUT_FILE}"

echo "Rendered ${OUTPUT_FILE}"
echo "Sources:"
echo "  ${COMMON_FILE}"
echo "  ${ENV_FILE}"
echo
echo "Preview:"
grep -E "^(APP_ENV|APP_DOMAIN|PORT|DATABASE_PATH|JWT_ISSUER|DEV_SEED_ADMIN|DEV_ADMIN_EMAIL|LLM_COACHING_ENABLED|VITE_API_BASE_URL|CORS_ALLOW_ORIGINS)=" "${OUTPUT_FILE}" || true

chmod +x scripts/render-env.sh
```

# Render All Environments

`scripts/render-all-envs.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

./scripts/render-env.sh dev .env.dev
./scripts/render-env.sh tst .env.tst
./scripts/render-env.sh prd .env.prd

# Convenience files for local development.
./scripts/render-env.sh dev backend/.env
./scripts/render-env.sh dev frontend/.env

echo
echo "Rendered all env files."

chmod +x scripts/render-all-envs.sh
```

**Run**:

`./scripts/render-all-envs.sh`

**This creates**:

```
.env.dev
.env.tst
.env.prd
backend/.env
frontend/.env
```

# Inspect Merged Files

`scripts/print-env.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-dev}"

TMP_FILE="$(mktemp)"
trap 'rm -f "${TMP_FILE}"' EXIT

./scripts/render-env.sh "${ENV_NAME}" "${TMP_FILE}" >/dev/null

cat "${TMP_FILE}"

chmod +x scripts/print-env.sh
```

**Use it**:

```bash
./scripts/print-env.sh dev
./scripts/print-env.sh prd
```

# Update Backend Environment
If present

`scripts/dev-backend.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND_DIR="${PROJECT_ROOT}/backend"

cd "${PROJECT_ROOT}"

if [[ ! -f "backend/.env" ]]; then
  ./scripts/render-env.sh dev backend/.env
fi

set -a
source backend/.env
set +a

cd "${BACKEND_DIR}"

mkdir -p data

echo "Starting backend..."
echo "APP_ENV=${APP_ENV:-dev}"
echo "PORT=${PORT:-8080}"
echo "DATABASE_PATH=${DATABASE_PATH:-data/app.db}"
echo "DEV_SEED_ADMIN=${DEV_SEED_ADMIN:-}"
echo "DEV_ADMIN_EMAIL=${DEV_ADMIN_EMAIL:-}"
echo "LLM_COACHING_ENABLED=${LLM_COACHING_ENABLED:-}"

go run ./cmd/api

chmod +x scripts/dev-backend.sh
```
# Update Frontend Environment
If present

`scripts/dev-frontend.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"

cd "${PROJECT_ROOT}"

if [[ ! -f "frontend/.env" ]]; then
  ./scripts/render-env.sh dev frontend/.env
fi

cd "${FRONTEND_DIR}"

npm run dev

chmod +x scripts/dev-frontend.sh
```

# Update production scripts to use rendered env files

## scripts/prod-build.sh
```bash
#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-prd}"
ENV_FILE=".env.${ENV_NAME}"

if [[ ! -f "${ENV_FILE}" ]]; then
  ./scripts/render-env.sh "${ENV_NAME}" "${ENV_FILE}"
fi

docker compose --env-file "${ENV_FILE}" -f docker-compose.prod.yml build

chmod +x scripts/prod-build.sh
```

## scripts/prod-up.sh
```bash
#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-prd}"
ENV_FILE=".env.${ENV_NAME}"

if [[ ! -f "${ENV_FILE}" ]]; then
  ./scripts/render-env.sh "${ENV_NAME}" "${ENV_FILE}"
fi

docker compose --env-file "${ENV_FILE}" -f docker-compose.prod.yml up -d

chmod +x scripts/prod-up.sh
```

## scripts/prod-down.sh
```bash
#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-prd}"
ENV_FILE=".env.${ENV_NAME}"

docker compose --env-file "${ENV_FILE}" -f docker-compose.prod.yml down

chmod +x scripts/prod-down.sh
```

## scripts/prod-logs.sh
```bash
#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-prd}"
ENV_FILE=".env.${ENV_NAME}"

docker compose --env-file "${ENV_FILE}" -f docker-compose.prod.yml logs -f --tail=200

chmod +x scripts/prod-logs.sh
```

# Usage
## Local development

**Render dev env files**:
```bash
./scripts/render-all-envs.sh
```

**Start backend**:
```bash
./scripts/dev-backend.sh
```

**Start frontend**:
```bash
./scripts/dev-frontend.sh
```

## Production simulation
```bash
./scripts/render-env.sh prd .env.prd
./scripts/prod-build.sh prd
./scripts/prod-up.sh prd
```

## Test/staging
```bash
./scripts/render-env.sh tst .env.tst
./scripts/prod-build.sh tst
./scripts/prod-up.sh tst
```

# Update .gitignore

**Add**:

```
.env
.env.*
backend/.env
frontend/.env

!backend/.env.example
!frontend/.env.example
!backend/.env.production.example
!frontend/.env.production.example
!env/common.env
!env/dev.env
!env/tst.env
!env/prd.env
```

**If you do not want to commit real tst.env or prd.env, use templates instead**:

```
env/tst.env.example
env/prd.env.example
```

**and ignore**:

```
env/tst.env
env/prd.env
```

For your current development stage, I would commit env/common.env and env/dev.env, but keep production secrets out of Git.

# Important naming note

Your current Go config expects:

```bash
DEV_ADMIN_EMAIL
DEV_ADMIN_PASSWORD
DEV_ADMIN_NAME
```

So I used those names.

Earlier you asked about:

```
ADMIN_EMAIL
ADMIN_PASSWORD
ADMIN_FULL_NAME
```

Those names are fine too, but only if your Go config.Load() reads them. To avoid breakage, use the names your backend already expects:

```
DEV_ADMIN_EMAIL
DEV_ADMIN_PASSWORD
DEV_ADMIN_NAME
```

The generated files now line up with the current backend code.