- [1. Environment Model](#1-environment-model)
- [2. Required Makefile Philosophy](#2-required-makefile-philosophy)
- [3. Recommended Makefile Target Map](#3-recommended-makefile-target-map)
- [4. Universal Repository Checklist](#4-universal-repository-checklist)
- [5. Local Environment Checklist](#5-local-environment-checklist)
  - [5.1 Verify local tools](#51-verify-local-tools)
- [5.2 Initialize local environment files](#52-initialize-local-environment-files)
  - [5.3 Initialize local database](#53-initialize-local-database)
  - [5.4 Run backend checks](#54-run-backend-checks)
  - [5.5 Run frontend checks](#55-run-frontend-checks)
  - [5.6 Run all local checks](#56-run-all-local-checks)
  - [5.7 Start local backend](#57-start-local-backend)
  - [5.8 Start local frontend](#58-start-local-frontend)
  - [5.9 Start local full stack](#59-start-local-full-stack)
- [5.10 Run local smoke tests](#510-run-local-smoke-tests)
  - [5.11 Local browser validation](#511-local-browser-validation)
  - [5.12 Local iPhone/LAN validation](#512-local-iphonelan-validation)
- [6. Development Environment Checklist](#6-development-environment-checklist)
  - [6.1 Initialize development env](#61-initialize-development-env)
  - [6.2 Build development containers](#62-build-development-containers)
  - [6.3 Start development environment](#63-start-development-environment)
  - [6.4 Check development containers](#64-check-development-containers)
  - [6.5 View development logs](#65-view-development-logs)
  - [6.6 Run development smoke tests](#66-run-development-smoke-tests)
  - [6.7 Development acceptance checklist](#67-development-acceptance-checklist)
- [7. Test/Staging Environment Checklist](#7-teststaging-environment-checklist)
  - [7.1 Direct provider setup exceptions](#71-direct-provider-setup-exceptions)
  - [7.2 Initialize staging env](#72-initialize-staging-env)
  - [7.3 Build staging](#73-build-staging)
  - [7.4 Start staging](#74-start-staging)
  - [7.5 Check staging containers](#75-check-staging-containers)
  - [7.6 Check staging certificate](#76-check-staging-certificate)
- [7.7 Run staging smoke tests](#77-run-staging-smoke-tests)
  - [7.8 Staging functional validation](#78-staging-functional-validation)
  - [7.9 Staging release gate](#79-staging-release-gate)
- [8. Production Environment Checklist](#8-production-environment-checklist)
  - [8.1 First-time server setup](#81-first-time-server-setup)
  - [8.2 GitHub deploy key setup](#82-github-deploy-key-setup)
  - [8.3 Clone repo](#83-clone-repo)
  - [8.4 Verify production DNS](#84-verify-production-dns)
  - [8.5 Initialize production env](#85-initialize-production-env)
  - [8.6 Build production containers](#86-build-production-containers)
  - [8.7 Start production](#87-start-production)
  - [8.8 Check logs](#88-check-logs)
  - [8.9 Production health checks](#89-production-health-checks)
  - [8.10 TLS certificate check](#810-tls-certificate-check)
  - [8.11 Run production smoke tests](#811-run-production-smoke-tests)
  - [8.12 Verify production admin login](#812-verify-production-admin-login)
  - [8.13 Disable admin seeding after bootstrap](#813-disable-admin-seeding-after-bootstrap)
  - [8.14 Production backup](#814-production-backup)
- [9. Environment Comparison Matrix](#9-environment-comparison-matrix)
- [10. Final Local Checklist](#10-final-local-checklist)
- [11. Final Development Checklist](#11-final-development-checklist)
- [12. Final Staging Checklist](#12-final-staging-checklist)
- [13. Final Production Checklist](#13-final-production-checklist)
- [14. Key Operational Rules](#14-key-operational-rules)
- [15. Makefile Recommended use](#15-makefile-recommended-use)
  - [For local development:](#for-local-development)
  - [For production on Hetzner:](#for-production-on-hetzner)
  - [For production admin verification:](#for-production-admin-verification)
# 1. Environment Model

Call It Cure It uses four operating environments:
| Environment | Description |
|-------------|-------------------------------------------------------|
| Local 	  |Developer laptop, fast iteration, two terminals |
| Development |Shared/dev deployment or containerized dev environment |
| Staging     |Production-like pre-release environment |
| Production  |Public Hetzner deployment |

Current application model:

Backend:
- Go 1.26 + Fiber
- SQLite/GORM
- JWT auth
- Admin user seeded by API startup through EnsureDevAdmin()
- No cmd/create-admin
- Optional LLM coaching
- Dockerized for deployed environments

**Frontend**:
- React + Vite + TypeScript
- Bootstrap
- React Query
- React Router
- VITE_API_BASE_URL=/api/v1

**Production**:
- Docker Compose
- Caddy reverse proxy
- HTTPS at https://app.callitcureit.com
- SQLite persisted in Docker volume

**Important architectural rules**:

- Do not use backend/cmd/create-admin.
- Do not build create-admin in Docker.
- The backend API seeds the admin user on startup when DEV_SEED_ADMIN=true.
- The frontend should use /api/v1, not localhost or LAN IPs.
- Production traffic should go through Caddy.

# 2. Required Makefile Philosophy

The Makefile should be the primary interface for developers and system operators.

A new developer should be able to run:

make local-setup
make local-check
make local-up

A systems engineer should be able to run:

make prod-init-env
make prod-build
make prod-up
make prod-smoke

Direct Docker, Go, npm, sqlite, and curl commands should be hidden behind Makefile targets wherever practical.

# 3. Recommended Makefile Target Map

Your root Makefile should expose targets similar to these:

```
General:
  make help
  make check
  make clean

Local:
  make local-init-env
  make local-db-init
  make local-db-reset
  make local-backend
  make local-frontend
  make local-up
  make local-check
  make local-smoke
  make local-login-test
  make local-admin-test

Development:
  make dev-init-env
  make dev-build
  make dev-up
  make dev-down
  make dev-logs
  make dev-smoke

Staging:
  make staging-init-env
  make staging-build
  make staging-up
  make staging-down
  make staging-logs
  make staging-smoke

Production:
  make prod-init-env
  make prod-build
  make prod-up
  make prod-down
  make prod-logs
  make prod-ps
  make prod-smoke
  make prod-health
  make prod-admin-test
  make prod-backup
  make prod-restart
  make prod-caddy-logs
  make prod-backend-logs
  make prod-frontend-logs

Diagnostics:
  make docker-ps
  make docker-df
  make prod-cert-check
  make prod-dns-check
  make prod-backend-health
  make prod-caddy-health
```

# 4. Universal Repository Checklist

Before any environment work, confirm the repository has the expected deployment files.

Run:

```bash
make check-repo
```

**This target should verify that these files exist**:

```bash
backend/
  Dockerfile
  docker-entrypoint.sh
  .dockerignore
  .env.example
  .env.production.example
  migrations/
  cmd/
    api/
      main.go

frontend/
  Dockerfile
  nginx.conf
  .dockerignore
  .env.example
  .env.production.example
  vite.config.ts

deploy/
  Caddyfile

scripts/
  init-dev-env.sh
  init-prod-env.sh
  dev-backend.sh
  dev-frontend.sh
  prod-build.sh
  prod-up.sh
  prod-down.sh
  prod-logs.sh
  prod-smoke-test.sh
  prod-backup-sqlite.sh

docker-compose.dev.yml
docker-compose.prod.yml
Makefile
```

**It should also fail if these exist**:

```bash
backend/cmd/create-admin
backend/cmd/create-admin.disabled
```

because create-admin is obsolete.

# 5. Local Environment Checklist

Use this for day-to-day development on a developer machine.

## 5.1 Verify local tools

**Run**:

```bash
make doctor
```

**This should verify**:

```bash
go
node
npm
sqlite3
docker
docker compose
jq
git
```

**Recommended versions**:

```
Go 1.26
Node 22+
SQLite CLI
Docker + Compose plugin
```

# 5.2 Initialize local environment files

**Run**:

```bash
make local-init-env
```

**This should create or update**:

```bash
backend/.env
frontend/.env
```

**Expected backend local values**:

```
APP_ENV=local
PORT=8080
DATABASE_PATH=data/app.db

JWT_SECRET=dev-secret-change-me
JWT_ISSUER=call-it-cure-it
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@example.com
DEV_ADMIN_PASSWORD=admin123
DEV_ADMIN_NAME=Admin User

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,http://192.168.2.154:5173
```

**Expected frontend local value**:

```
VITE_API_BASE_URL=/api/v1
```

## 5.3 Initialize local database

**Run**:
```bash
make local-db-init
```

**This should**:
- create backend/data/
- create backend/data/app.db if needed
- apply all migrations in order
- verify expected tables exist

**Expected tables**:
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

**To reset local runtime/session data**:

```bash
make local-db-reset
```

**This should clear**:

```
session_scores
action_evaluations
trainee_actions
session_events
sessions
```

**To reset the local admin user**:

```bash
make local-admin-reset
```

Then restart the backend to reseed the admin user.

## 5.4 Run backend checks

**Run**:
```bash
make backend-check
```

**This should run**:
```bash
go mod tidy check
go test ./...
```

No direct go test command should be needed during normal use.

## 5.5 Run frontend checks

**Run**:
```bash
make frontend-check
```

**This should run**:
```bash
npm install if needed
npm run check
```

**The frontend check should include**:
```
typecheck
lint
build
```

## 5.6 Run all local checks

**Run**:
```bash
make local-check
```

**This should run**:
```
backend checks
frontend checks
repository checks
```

## 5.7 Start local backend

**Run**:
```bash
make local-backend
```

**This should start the backend with**:
```bash
DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@example.com
DEV_ADMIN_PASSWORD=admin123
DEV_ADMIN_NAME=Admin User
JWT_SECRET=dev-secret-change-me
DATABASE_PATH=data/app.db
```

**Expected backend output**:
```
API listening on :8080
Database path: data/app.db
Dev admin seed enabled: true
Dev admin email: admin@example.com
```

## 5.8 Start local frontend

**In another terminal, run**:
```bash
make local-frontend
```

**This should start Vite at**:
```
http://localhost:5173
```

**The frontend should use the Vite proxy to call the API through**:

`/api/v1`

## 5.9 Start local full stack

Use two terminals:
```bash
# terminal 1
make local-backend
# terminal 2
make local-frontend
```

# 5.10 Run local smoke tests

Run:
```bash
make local-smoke
```

**This should verify**:

```bash
http://localhost:8080/api/v1/healthz
http://localhost:5173/api/v1/healthz
http://localhost:5173/api/v1/scenarios
```

**Run login test**:
```bash
make local-login-test
```

**Run admin route test**:
```bash
make local-admin-test
```

**Expected**:
```
login returns JWT token
/admin/scenarios returns data with bearer token
```

## 5.11 Local browser validation

**Open**:

```bash
http://localhost:5173
```

**Checklist**:
```
[ ] /scenarios loads
[ ] scenario-hearsay-001 loads
[ ] training session starts
[ ] transcript advances
[ ] Objection, hearsay. produces Sustained.
[ ] Coach feedback appears
[ ] /login works
[ ] /admin/scenarios works
[ ] scenario edit page loads
[ ] objection types load
[ ] line opportunities render
```
## 5.12 Local iPhone/LAN validation

Direct command exception: find your Mac LAN IP.

```bash
ipconfig getifaddr en0
```

**Example**:
```
192.168.2.154
```

**Then on iPhone open**:
```
http://192.168.2.154:5173/login
```

**Do not use this on iPhone**:
```
http://localhost:5173
```

because localhost on the iPhone means the iPhone itself.

**Use Makefile proxy test**:
```bash
make local-lan-smoke LAN_HOST=192.168.2.154
```

**This should verify**:
```
http://192.168.2.154:5173/api/v1/healthz
```

**Remember**:
```
localhost:5173
192.168.2.154:5173
```

have separate browser storage. Log in separately on each origin.

# 6. Development Environment Checklist

Use this for shared development or containerized dev deployment.

## 6.1 Initialize development env

**Run**:
```bash
make dev-init-env
```

**This should create**:

```bash
.env.development
```

**Recommended values**:
```bash
APP_ENV=development
APP_DOMAIN=dev.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=development-long-random-secret
JWT_ISSUER=call-it-cure-it-dev
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@example.com
DEV_ADMIN_PASSWORD=admin123
DEV_ADMIN_NAME=Admin User

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://dev.callitcureit.com
```

## 6.2 Build development containers

**Run**:
```bash
make dev-build
```

## 6.3 Start development environment

**Run**:
```bash
make dev-up
```

## 6.4 Check development containers

**Run**:
```bash
make dev-ps
```

**Expected**:
```
backend running
frontend running if included
caddy running if included
```

## 6.5 View development logs

**Run**:
```bash
make dev-logs
```

**Backend logs only**:
```bash
make dev-backend-logs
```

## 6.6 Run development smoke tests

**Run**:
```bash
make dev-smoke
```

**This should verify**:
```
health endpoint
scenario endpoint
login endpoint
admin scenario endpoint
```

## 6.7 Development acceptance checklist
```
[ ] make dev-build succeeds
[ ] make dev-up starts containers
[ ] make dev-smoke passes
[ ] admin user is seeded
[ ] /scenarios loads
[ ] /login works
[ ] /admin/scenarios works
[ ] scenario editor loads
[ ] backend logs are clean
```
# 7. Test/Staging Environment Checklist

Use staging for production-like release validation.

## 7.1 Direct provider setup exceptions

Some staging tasks are still manual:

- create DNS record for staging.callitcureit.com
- generate or store staging secrets
- configure GitHub or deployment provider if needed

**DNS should point to the staging server**:

```
staging.callitcureit.com A <Hetzner staging-server-ip>
```

**Verify using Makefile**:
```bash
make staging-dns-check
```

## 7.2 Initialize staging env

**Run**:
```bash
make staging-init-env
```

**This should create**:

```
.env.staging
```

**Recommended staging value**s:
```bash
APP_ENV=staging
APP_DOMAIN=staging.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=<staging-long-random-secret>
JWT_ISSUER=call-it-cure-it-staging
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@callitcureit.com
DEV_ADMIN_PASSWORD=<temporary-staging-password>
DEV_ADMIN_NAME=Staging Admin

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://staging.callitcureit.com
```

Direct command exception: generate JWT secret if needed.

```bash
openssl rand -base64 48
```

Then edit .env.staging.

## 7.3 Build staging

**Run**:
```bash
make staging-build
```

## 7.4 Start staging

**Run**:
```bash
make staging-up
```

## 7.5 Check staging containers

**Run**:
```bash
make staging-ps
```

**Run logs**:
```bash
make staging-logs
```

## 7.6 Check staging certificate

**Run**:
```bash
make staging-cert-check
```

**Expected certificate SAN**:
```
DNS:staging.callitcureit.com
```

# 7.7 Run staging smoke tests

**Run**:
```bash
make staging-smoke
```

**This should verify**:
```
frontend responds
/api/v1/healthz responds
/api/v1/scenarios responds
login works
admin scenarios route works
```

## 7.8 Staging functional validation

**Use the browser**:
```
https://staging.callitcureit.com
```

**Checklist**:
```
[ ] frontend loads
[ ] /scenarios loads
[ ] scenario detail loads
[ ] training session starts
[ ] objection submission works
[ ] judge ruling appears
[ ] coach feedback appears
[ ] score/debrief works if enabled
[ ] /login works
[ ] /admin/scenarios works
[ ] edit scenario loads
[ ] objection types load
[ ] no major browser console errors
```
## 7.9 Staging release gate

**Before pro```duction**:
```
[ ] make backend-check passes
[ ] make frontend-check passes
[ ] make staging-build passes
[ ] make staging-up succeeds
[ ] make staging-smoke passes
[ ] make staging-cert-check passes
[ ] admin login works
[ ] training flow works
[ ] admin edit flow works
[ ] backup script works
[ ] logs are clean
```
# 8. Production Environment Checklist

Use this for the public Hetzner deployment.

**Production domain**:
```
app.callitcureit.com
```

**Production server**:
```
5.78.208.230
```
## 8.1 First-time server setup

Some commands must be run directly because the Makefile is not available until the repo is cloned.

Install server dependencies:
```bash
sudo apt update
sudo apt upgrade -y

sudo apt install -y \
  ca-certificates \
  curl \
  git \
  ufw \
  fail2ban \
  sqlite3 \
  jq \
  openssl
```

**Install Docker**:

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker deploy
```

**Reconnect**:
```bash
exit
ssh deploy@5.78.208.230
```

**Verify**:
```bash
docker ps
docker compose version
```

**Firewall**:
```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status
```

## 8.2 GitHub deploy key setup 

Direct command exception: generate SSH deploy key.
```bash
ssh-keygen -t ed25519 -C "hetzner-callitcureit-deploy" -f ~/.ssh/callitcureit_deploy
```

**Create SSH config**:
```bash
nano ~/.ssh/config
```

**Use**:
```
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/callitcureit_deploy
  IdentitiesOnly yes
```

**Permissions**:
```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/config
chmod 600 ~/.ssh/callitcureit_deploy
chmod 644 ~/.ssh/callitcureit_deploy.pub
```

**Copy the public key**:
```bash
cat ~/.ssh/callitcureit_deploy.pub
```

**Manual external step**:

```
GitHub repository → Settings → Deploy keys → Add deploy key
```

**Test**:
```bash
ssh -T git@github.com
```

## 8.3 Clone repo

Direct command exception because Makefile is not available yet.
```bash
mkdir -p /opt/callitcureit
cd /opt/callitcureit
git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git app
cd app
```

**From this point forward, use make**.

## 8.4 Verify production DNS

**Run**:
```bash
make prod-dns-check
```

**Expected**:
```
app.callitcureit.com A 5.78.208.230
```

**Manual external step if wrong**:

```
Update DNS provider:
app.callitcureit.com A 5.78.208.230
```

If an invalid A record exists, remove or correct it.

## 8.5 Initialize production env

**Run**:
```bash
make prod-init-env
```

**This should create**:
```
.env.production
```

**Then edit secrets manually**:
```bash
nano .env.production
```

**Expected production values**:
```bash
APP_ENV=production
APP_DOMAIN=app.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=<long-random-secret>
JWT_ISSUER=call-it-cure-it
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@callitcureit.com
DEV_ADMIN_PASSWORD=<temporary-strong-password>
DEV_ADMIN_NAME=Admin User

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://app.callitcureit.com
```

Direct command exception: generate a JWT secret if the script did not already generate one.
```bash
openssl rand -base64 48
```
Important:
```bash
DEV_SEED_ADMIN=true
```
only for first bootstrap. After successful admin login, set DEV_SEED_ADMIN=false.

## 8.6 Build production containers

**Run**:
```bash
make prod-build
```

**This should use**:
`
.env.production
docker-compose.prod.yml
BUILDX_NO_DEFAULT_ATTESTATIONS=1
--progress=plain
`

**If disk space is low, run**:
```bash
make docker-df
```

**Then, if needed**:
```bash
make docker-prune-build-cache
```

Avoid pruning volumes unless a database backup exists.

## 8.7 Start production

**Run**:
```bash
make prod-up
```

**Check**:
```bash
make prod-ps
```

**Expected**:
```
callitcureit-backend    running / healthy
callitcureit-frontend   running
callitcureit-caddy      running
```

## 8.8 Check logs

**Run**:
```bash
make prod-logs
```

**Backend only**:
```bash
make prod-backend-logs
```

**Frontend only**:
```bash
make prod-frontend-logs
```

**Caddy only**:
```bash
make prod-caddy-logs
```

**If backend says**:
```
no such table: users
```

then migrations did not run. 

**Confirm the image contains backend/docker-entrypoint.sh and rebuild:**
```bash
make prod-build
make prod-up
```

## 8.9 Production health checks

**Run backend internal health**:
```bash
make prod-backend-health
```

This should run health inside the backend container, because backend port 8080 is not exposed to the host in production.

**Run Caddy local health**:
```bash
make prod-caddy-health
```

**Run public health**:
```bash
make prod-health
```

**Expected**:
```
HTTP 200
{"status":"ok"}
```

## 8.10 TLS certificate check

**Run**:
```bash
make prod-cert-check
```

**Expected certificate SAN includes**:
```
DNS:app.callitcureit.com
```

**If this fails, check**:
```bash
make prod-dns-check
make prod-caddy-logs
```

Manual checks may be needed for DNS provider or firewall.

## 8.11 Run production smoke tests

**Run**:
```bash
make prod-smoke
```

This should verify:
```
frontend responds
/api/v1/healthz responds
/api/v1/scenarios responds
login works if credentials are configured
admin scenarios works if token is available
```

## 8.12 Verify production admin login

**Run**:
```bash
make prod-admin-test
```

**This should log in with**:
```
DEV_ADMIN_EMAIL
DEV_ADMIN_PASSWORD
```

**from .env.production, then call**:
```
/api/v1/admin/scenarios
```

**Also verify in browser**:
```
https://app.callitcureit.com/login
https://app.callitcureit.com/admin/scenarios
```

## 8.13 Disable admin seeding after bootstrap

**After production admin login works, edit**:
```bash
nano .env.production
```
**Change**:
```
DEV_SEED_ADMIN=false
```
**Restart**:
```bash
make prod-up
```
**Verify**:
```bash
make prod-smoke
make prod-admin-test
```

## 8.14 Production backup

**Run**:
```bash
make prod-backup
```
**Expected backup**:
```
backups/app-YYYYMMDD-HHMMSS.db
```

**Copy backups off server regularly.**

**Direct command exception**:

```bash
scp deploy@5.78.208.230:/opt/callitcureit/app/backups/app-*.db ./backups/
```bash

## 8.15 Production update procedure

**For every release**:
```bash
make prod-update
```
**This should perform**:
```bash
git pull
git checkout production
prod-build
prod-up
prod-smoke
```

**If you prefer explicit steps**:
```bash
git pull
git checkout production
make prod-build
make prod-up
make prod-smoke
```

Direct git pull is acceptable because it interacts with the repository state, but you may wrap it in make prod-update.

# 9. Environment Comparison Matrix
| Task | Local	| Development |	Staging	| Production |
|------|--------|-------------|---------|------------|  
| Initialize env	| make local-init-env	| make dev-init-env	| make staging-init-env	| make prod-init-env |  
| Build | make local-check | make dev-build | make staging-build | make prod-build | 
| Start | make local-up | make dev-up | make staging-up | make prod-up | 
| Logs | terminal | make local-logs | make dev-logs | make staging-logs | make prod-logs | 
| Health | make | local-smoke | make dev-smoke | make staging-smoke | make prod-smoke | 
| TLS | one	optional | make staging-cert-check | make prod-cert-check | 
| Backup | optional | optional | recommended	make prod-backup | 
| Admin | seed | true | true | true initially | true initially, then false | 

# 10. Final Local Checklist
```
[ ] make doctor passes
[ ] make local-init-env creates backend/.env and frontend/.env
[ ] make local-db-init creates SQLite database
[ ] make backend-check passes
[ ] make frontend-check passes
[ ] make local-backend starts API
[ ] make local-frontend starts Vite
[ ] make local-smoke passes
[ ] make local-login-test passes
[ ] make local-admin-test passes
[ ] /scenarios works in browser
[ ] /login works in browser
[ ] /admin/scenarios works in browser
[ ] scenario edit page works
[ ] training flow works
```
# 11. Final Development Checklist
```
[ ] make dev-init-env completed
[ ] make dev-build succeeds
[ ] make dev-up starts containers
[ ] make dev-ps shows expected services
[ ] make dev-smoke passes
[ ] admin user seeded
[ ] public scenario flow works
[ ] admin flow works
[ ] backend logs clean
```
# 12. Final Staging Checklist
```
[ ] staging DNS configured
[ ] make staging-init-env completed
[ ] staging secrets edited
[ ] make staging-build succeeds
[ ] make staging-up starts stack
[ ] make staging-cert-check passes
[ ] make staging-smoke passes
[ ] admin login works
[ ] scenario training flow works
[ ] admin scenario edit works
[ ] backup works
[ ] release approved for production
```
# 13. Final Production Checklist
```
[ ] Hetzner server provisioned
[ ] Docker installed
[ ] deploy user is in docker group
[ ] UFW allows OpenSSH, 80, 443
[ ] GitHub deploy key configured
[ ] repo cloned to /opt/callitcureit/app
[ ] app.callitcureit.com DNS points to 5.78.208.230
[ ] make prod-init-env completed
[ ] .env.production edited with real values
[ ] JWT_SECRET is strong
[ ] DEV_ADMIN_PASSWORD is strong
[ ] make prod-build succeeds
[ ] make prod-up starts stack
[ ] make prod-ps shows backend, frontend, caddy
[ ] make prod-backend-health passes
[ ] make prod-cert-check passes
[ ] make prod-smoke passes
[ ] make prod-admin-test passes
[ ] DEV_SEED_ADMIN changed to false
[ ] make prod-up restarted stack after disabling seed
[ ] make prod-backup creates SQLite backup
```
# 14. Key Operational Rules
- Use make for normal operations.
- Do not use cmd/create-admin.
- Do not expose backend :8080 publicly in production.
- Do not set frontend production API URL to localhost.
- Use VITE_API_BASE_URL=/api/v1.
- Use Caddy for public HTTPS.
- Use DEV_SEED_ADMIN=true only for initial bootstrap.
- Set DEV_SEED_ADMIN=false after successful production admin login.
- Back up SQLite regularly.
- Do not commit .env.production.

# 15. Makefile Recommended use
## For local development:
```bash
# terminal 1
make init-dev
make db-init
make check
make dev-backend
```
```bash
# terminal 2
make dev-frontend
```
## For production on Hetzner:
```bash
make init-prod
nano .env.production
make docker-prod-build
make docker-prod-up
make docker-prod-ps
make prod-smoke BASE_URL=https://app.callitcureit.com
```
## For production admin verification:
```bash
ADMIN_EMAIL=admin@callitcureit.com \
ADMIN_PASSWORD='your-production-password' \
BASE_URL=https://app.callitcureit.com \
make prod-admin-scenarios-test
```