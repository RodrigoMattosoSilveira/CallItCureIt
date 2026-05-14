# Required local tools

## Install:

```
Go
Node.js + npm
SQLite CLI
Docker Desktop or Docker Engine
jq
Git
```

## Verify:

```
go version
node --version
npm --version
sqlite3 --version
docker --version
jq --version
```

# Clone repository

```bash
git clone <repo-url> CallItCureIt
cd CallItCureIt
```

## Expected structure:

```bash
CallItCureIt/
  backend/
  frontend/
  scripts/
  docker-compose.dev.yml
  docker-compose.prod.yml
```

# Backend environment setup

## Create backend .env from example if present:

```bash
cp backend/.env.example backend/.env
```

## For local development, make sure these values exist:

```
PORT=8080
DATABASE_PATH=data/app.db

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@example.com
DEV_ADMIN_PASSWORD=admin123
DEV_ADMIN_NAME=Admin User

JWT_SECRET=dev-secret-change-me
JWT_ISSUER=call-it-cure-it
JWT_EXPIRATION_MINUTES=480

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173
```

## If using your env script:

```bash
chmod +x scripts/init-env.sh
./scripts/init-env.sh
```

# Create local SQLite database

## From project root:

```bash
mkdir -p backend/data
```

## Apply migrations in order:

```bash
sqlite3 backend/data/app.db < backend/migrations/000001_init_schema.up.sql
sqlite3 backend/data/app.db < backend/migrations/000002_seed_reference_data.up.sql
sqlite3 backend/data/app.db < backend/migrations/000003_create_sessions.up.sql
sqlite3 backend/data/app.db < backend/migrations/000004_create_trainee_actions.up.sql
sqlite3 backend/data/app.db < backend/migrations/000005_create_action_evaluations.up.sql
sqlite3 backend/data/app.db < backend/migrations/000006_update_session_scores.up.sql
sqlite3 backend/data/app.db < backend/migrations/000007_create_users.up.sql
```

## Verify tables:

```bash
sqlite3 backend/data/app.db ".tables"
```

## Expected tables include:

```
scenarios
scenario_lines
scenario_actors
objection_types
objection_opportunities
rule_refs
sessions
session_events
trainee_actions
action_evaluations
session_scores
users
```

## Verify seeded scenario:

```bash
sqlite3 backend/data/app.db "
SELECT id, title, status FROM scenarios;
"
```

## Expected:

`scenario-hearsay-001|Basic Hearsay on Direct Examination|published`

# Backend dependency setup

```bash
cd backend
go mod tidy
go test ./...
```

## Expected:

`PASS`

If cmd/create-admin still exists and fails, remove or move it outside backend/cmd because admin is now seeded by API startup:

```bash
mkdir -p ../disabled
mv cmd/create-admin ../disabled/create-admin
```

## Then rerun:

```bash
go test ./...
```

# Start backend locally

## Preferred:

```bash
cd backend

DEV_SEED_ADMIN=true \
DEV_ADMIN_EMAIL=admin@example.com \
DEV_ADMIN_PASSWORD=admin123 \
DEV_ADMIN_NAME='Admin User' \
JWT_SECRET=dev-secret-change-me \
DATABASE_PATH=data/app.db \
go run ./cmd/api
```

## Or use:

```bash
./scripts/dev-backend.sh
```

## Expected logs:

```
API listening on :8080
Database path: data/app.db
Dev admin seed enabled: true
Dev admin email: admin@example.com
```

# Backend smoke tests

## In another terminal:

```bash
curl -i http://localhost:8080/api/v1/healthz
```

## Expected:

```
{"status":"ok"}
```

## Test scenarios:

```bash
curl -s http://localhost:8080/api/v1/scenarios | jq
```

## Test login:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }' | jq
```

## Expected:

```json
{
  "data": {
    "token": "...",
    "user": {
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

## Test protected admin route:

```bash
TOKEN=$(
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "admin@example.com",
      "password": "admin123"
    }' | jq -r '.data.token'
)

curl -s http://localhost:8080/api/v1/admin/scenarios \
  -H "Authorization: Bearer $TOKEN" | jq
```

# Frontend environment setup

## Create:

`frontend/.env`

## Use this for local development:

```
VITE_API_BASE_URL=/api/v1
```

This works with the Vite proxy and avoids switching between localhost and LAN IP.

# Confirm Vite proxy

## Your frontend/vite.config.ts should include:

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

```json
export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
```

# Frontend dependency setup

```bash
cd frontend
npm install
npm run check
```

## Expected:

```
typecheck passes
lint passes
build passes
```

# Start frontend locally

```bash
cd frontend
npm run dev
```

Or:

```bash
./scripts/dev-frontend.sh
```

## Open on local machine:

`http://localhost:5173`

# Local browser test flow

## Open:

`http://localhost:5173/scenarios`

## Expected:

`Basic Hearsay on Direct Examination`

## Test training flow:

1. Open scenario.
2. Start training session.
3. Click Next Line until line 4.
4. Submit: Objection, hearsay.
5. Confirm judge says: Sustained.
6. Confirm coach feedback appears.
7. Confirm score/debrief behavior if implemented.

##Test admin flow:

1. Go to /login.
2. Log in as admin@example.com / admin123.
3. Confirm redirect to /admin/scenarios.
4. Click Edit.
5. Confirm scenario detail loads.

# iPhone or LAN testing checklist

# Find Mac LAN IP:

`ipconfig getifaddr en0``

Example:

`192.168.2.154`

## Backend should be running on:

`localhost:8080`

Frontend should be running with Vite host:

`0.0.0.0:5173`

## Open on iPhone:

`http://192.168.2.154:5173/login`

## Important:

Do not use localhost on iPhone.
localhost on iPhone means the iPhone itself.

## Confirm proxy from Mac:

`curl -i http://192.168.2.154:5173/api/v1/healthz`

If this fails, the iPhone will fail too.

If login fails on iPhone:

1. Clear Safari website data for 192.168.2.154.
2. Reopen http://192.168.2.154:5173/login.
3. Log in again.

## Remember:

`localhost:5173 and 192.168.2.154:5173 have separate localStorage.`

# Useful local reset commands

Reset session/test data:

```bash
sqlite3 backend/data/app.db "
DELETE FROM session_scores;
DELETE FROM action_evaluations;
DELETE FROM trainee_actions;
DELETE FROM session_events;
DELETE FROM sessions;
"
```

## Reset admin user:

```bash
sqlite3 backend/data/app.db "
DELETE FROM users WHERE email = 'admin@example.com';
"
```

Restart backend to reseed admin.