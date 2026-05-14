# Systems Engineer Production Deployment Checklist for Hetzner
This section assumes a Hetzner Ubuntu server with Docker Compose and Caddy.

# Recommended production topology:

```
Internet
  ↓
Caddy on ports 80/443
  ↓
Frontend container
  ↓ same origin /api/v1
Backend container
  ↓
SQLite volume
```

## Production URLs:

```
https://app.callitcureit.com/          -> frontend
https://app.callitcureit.com/api/v1/*  -> backend
```
# Prepare Hetzner server
## Create server

**Recommended minimum for MVP:**

```
Ubuntu 24.04 LTS
2 vCPU
4 GB RAM
40+ GB disk
```

For production with LLM usage, monitoring, and backups, prefer more disk and RAM.

## Add SSH key

**From local machine**:

```bash
ssh root@YOUR_SERVER_IP
```

**Create deploy user**:

```bash
adduser deploy
usermod -aG sudo deploy
```

**Add SSH key**:

```bash
mkdir -p /home/deploy/.ssh
cp /root/.ssh/authorized_keys /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
```

**Log in as deploy**:

```bash
ssh deploy@YOUR_SERVER_IP
```

# Basic server hardening

**Update packages**:

```bash
sudo apt update
sudo apt upgrade -y
```

**Install basics**:

```bash
sudo apt install -y \
  ca-certificates \
  curl \
  git \
  ufw \
  fail2ban \
  sqlite3 \
  jq
```

**Enable firewall**:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status
```

## Optiona
disable root SSH login after confirming deploy user works.

**Edit**:

```bash
sudo nano /etc/ssh/sshd_config
```

**Set**:

```
PermitRootLogin no
PasswordAuthentication no
```

**Restart SSH**:

```bash
sudo systemctl restart ssh
```

# Install Docker

**Install Docker Engine and Compose plugin**:

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker deploy
```

**Log out and back in**.

**Verify**:

```bash
docker --version
docker compose version
```

# DNS setup

**Create DNS A record**:

```
app.callitcureit.com -> YOUR_HETZNER_SERVER_IP
```

**Verify from your local machine**:

```bash
dig app.callitcureit.com
```

**or**:

```bash
nslookup app.callitcureit.com
```

Wait until it resolves to the Hetzner IP.

# Clone application on server

**On Hetzner**:

```bash
mkdir -p /opt/callitcureit
sudo chown -R deploy:deploy /opt/callitcureit
cd /opt/callitcureit

git clone <repo-url> app
cd app
```

# Production environment file

**Create**:

```bash
cp backend/.env.production.example .env.production
nano .env.production
```

**Recommended production values**:

```
APP_DOMAIN=app.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=replace-with-a-long-random-secret
JWT_ISSUER=call-it-cure-it
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@callitcureit.com
DEV_ADMIN_PASSWORD=temporary-strong-password
DEV_ADMIN_NAME=Admin User

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://app.callitcureit.com
```

**Generate a strong JWT secret**:

```bash
openssl rand -base64 48
```

**Use the output as**:

```
JWT_SECRET=<generated-secret>
```

Important:

```
DEV_SEED_ADMIN=true is acceptable for first deployment only.
After admin is created and verified, set DEV_SEED_ADMIN=false and redeploy.
```

# Confirm production packaging files exist

**Required**:

```
backend/Dockerfile
frontend/Dockerfile
frontend/nginx.conf
docker-compose.prod.yml
deploy/Caddyfile
```

**Required scripts**:

```
scripts/prod-build.sh
scripts/prod-up.sh
scripts/prod-down.sh
scripts/prod-logs.sh
scripts/prod-smoke-test.sh
scripts/prod-backup-sqlite.sh
```

**Make scripts executable**:

```bash
chmod +x scripts/*.sh
```

# Build production containers

```bash
./scripts/prod-build.sh .env.production
```

**Or directly**:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml build
```bash

# Start production stack

```bash
./scripts/prod-up.sh .env.production
```

**Or directly**:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml up -d
```

**Check containers**:

```bash
docker ps
```

**Expected**:

```
callitcureit-backend
callitcureit-frontend
callitcureit-caddy
```

**View logs**:

```bash
./scripts/prod-logs.sh .env.production
```

# Production smoke tests

**From server**:

```bash
BASE_URL=https://app.callitcureit.com ./scripts/prod-smoke-test.sh
```

**Or manually**:

```bash
curl -i https://app.callitcureit.com/
curl -i https://app.callitcureit.com/api/v1/healthz
curl -s https://app.callitcureit.com/api/v1/scenarios | jq
```

**Expected health**:

```json
{"status":"ok"}
```

# Verify admin login

**From local browser**:

`https://app.callitcureit.com/login`

**Login**:

```
admin@callitcureit.com
temporary-strong-password
```

**Verify admin scenarios**:

`https://app.callitcureit.com/admin/scenarios`

**Or with curl**:

```bash
TOKEN=$(
  curl -s -X POST https://app.callitcureit.com/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "admin@callitcureit.com",
      "password": "temporary-strong-password"
    }' | jq -r '.data.token'
)

curl -s https://app.callitcureit.com/api/v1/admin/scenarios \
  -H "Authorization: Bearer $TOKEN" | jq
```

# Disable production dev-admin seeding

**After first successful login**:

**Edit**:

```bash
nano .env.production
```

**Change**:

```
DEV_SEED_ADMIN=false
```

**Then restart**:

```bash
./scripts/prod-up.sh .env.production
```

**Confirm app still works**:

```bash
BASE_URL=https://app.callitcureit.com ./scripts/prod-smoke-test.sh
```

# SQLite persistence checklist

In production, SQLite must live in a Docker volume:

```yaml
volumes:
  backend-data:
```

**Backend uses**:

```
DATABASE_PATH=/app/data/app.db
```

**Confirm DB exists inside container**:

```bash
docker exec -it callitcureit-backend ls -l /app/data
```

**Backup**:

```bash
./scripts/prod-backup-sqlite.sh
```

**Expected**:

```
backups/app-YYYYMMDD-HHMMSS.db
```

**Copy backups off server regularly**:

```bash
scp deploy@app.callitcureit.com:/opt/callitcureit/app/backups/app-*.db ./backups/
```

# Restore SQLite backup

**Stop stack**:

```bash
./scripts/prod-down.sh .env.production
```

**Find Docker volume mount**:

```bash
docker volume ls
docker volume inspect app_backend-data
```

**Simpler restore method**:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml up -d backend
docker cp backups/app-YYYYMMDD-HHMMSS.db callitcureit-backend:/app/data/app.db
docker compose --env-file .env.production -f docker-compose.prod.yml restart backend
```

**Then restart full stack**:

```bash
./scripts/prod-up.sh .env.production
```

# Log management

**View all logs**:

```bash
./scripts/prod-logs.sh .env.production
```

**Backend only**:

```bash
docker logs -f --tail=200 callitcureit-backend
```

**Caddy only**:

```bash
docker logs -f --tail=200 callitcureit-caddy
```

**Frontend only**:

```bash
docker logs -f --tail=200 callitcureit-frontend
```

# Production update procedure

**On server**:

```bash
cd /opt/callitcureit/app
git pull
./scripts/prod-build.sh .env.production
./scripts/prod-up.sh .env.production
BASE_URL=https://app.callitcureit.com ./scripts/prod-smoke-test.sh
```

If migration files changed, confirm your backend has a migration strategy. If migrations are still manual, run them before restart or add an entrypoint migration step.

**Manual migration example**:

```bash
docker exec -i callitcureit-backend sqlite3 /app/data/app.db < backend/migrations/000008_some_migration.up.sql
```

A better production approach is to add a controlled migration runner later.

# Caddy/TLS checklist

**Confirm ports open**:

```bash
sudo ufw status
```

Should include:

```
80/tcp
443/tcp
OpenSSH
```

**Confirm Caddy got a certificate**:

```bash
docker logs callitcureit-caddy
```

*Look for successful certificate issuance*.

**Test**:

```bash
curl -I https://app.callitcureit.com
```

**Expected**:

```
HTTP/2 200
```

# CORS checklist

**Production should use same origin**:

```
https://app.**callitcureit.com**
```

**Frontend should use**:

```
VITE_API_BASE_URL=/api/v1
```

**Backend should allow**:

```
CORS_ALLOW_ORIGINS=https://app.callitcureit.com
```

Because Caddy routes /api/v1 to backend, CORS should rarely be an issue in production.

# LLM/OpenAI production checklist

**If enabling LLM coaching**:

```
LLM_COACHING_ENABLED=true
OPENAI_API_KEY=<production-secret>
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20
```

**Then redeploy**:

```bash
./scripts/prod-up.sh .env.production
```

**Never commit**:

```
OPENAI_API_KEY
JWT_SECRET
admin password
```

# Security checklist

**Before exposing publicly**:

```
Use strong JWT_SECRET.
Change temporary admin password.
Set DEV_SEED_ADMIN=false after first setup.
Use HTTPS only.
Keep port 8080 private inside Docker network.
Only expose 80/443 publicly.
Keep .env.production out of Git.
Back up SQLite database.
Limit SSH to key-based auth.
Disable root SSH login.
Enable firewall.
Enable fail2ban.
```

# Developer handoff checklist

A developer should be able to run:

```bash
./scripts/dev-backend.sh
./scripts/dev-frontend.sh
```

Then open:

```
http://localhost:5173
```

A systems engineer should be able to run:

```bash
./scripts/init-env.sh
./scripts/prod-build.sh .env.production
./scripts/prod-up.sh .env.production
BASE_URL=https://app.callitcureit.com ./scripts/prod-smoke-test.sh
```

# Troubleshooting quick reference

## Frontend cannot load scenarios

**Check**:

```bash
curl -i http://localhost:5173/api/v1/healthz
curl -i http://localhost:8080/api/v1/healthz
```

## Admin route returns 401

**Check token**:

```
localStorage.getItem("auth_token")
```

Log in again.

## Admin route returns 404

Backend route is missing. Check cmd/api/main.go admin route registration.

## iPhone cannot log in

**Use LAN URL**:

```
http://YOUR_MAC_IP:5173/login
```

**Do not use**:

```
http://localhost:5173
```

## Production TLS fails

**Check**:

```bash
docker logs callitcureit-caddy
dig app.callitcureit.com
sudo ufw status
SQLite schema error
```

**Check**:

```bash
docker exec -it callitcureit-backend sqlite3 /app/data/app.db ".schema"
```

# Final recommended sequence
## Developer
```bash
git clone <repo-url> CallItCureIt
cd CallItCureIt
./scripts/init-env.sh backend/.env.example backend/.env
mkdir -p backend/data
sqlite3 backend/data/app.db < backend/migrations/000001_init_schema.up.sql
sqlite3 backend/data/app.db < backend/migrations/000002_seed_reference_data.up.sql
sqlite3 backend/data/app.db < backend/migrations/000003_create_sessions.up.sql
sqlite3 backend/data/app.db < backend/migrations/000004_create_trainee_actions.up.sql
sqlite3 backend/data/app.db < backend/migrations/000005_create_action_evaluations.up.sql
sqlite3 backend/data/app.db < backend/migrations/000006_update_session_scores.up.sql
sqlite3 backend/data/app.db < backend/migrations/000007_create_users.up.sql
cd backend && go test ./...
cd ../frontend && npm install && npm run check
cd ..
./scripts/dev-backend.sh
./scripts/dev-frontend.sh
```

## Systems engineer
```bash
ssh deploy@app.callitcureit.com
cd /opt/callitcureit/app
cp backend/.env.production.example .env.production
nano .env.production
./scripts/prod-build.sh .env.production
./scripts/prod-up.sh .env.production
BASE_URL=https://app.callitcureit.com ./scripts/prod-smoke-test.sh
./scripts/prod-backup-sqlite.sh
```bash

This gives you a complete local and production checklist aligned with the current application state.