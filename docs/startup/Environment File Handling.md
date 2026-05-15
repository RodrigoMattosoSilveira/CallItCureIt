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

# Usage
## Local development

**Render local development env files**:

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

**Local URLs**:

```
http://localhost:5173
http://localhost:5173/login
http://localhost:5173/admin/scenarios
```

**For iPhone/LAN testing**:

```
http://192.168.2.154:5173
http://192.168.2.154:5173/login
```

# Production simulation

Use this when you want to test the production Docker setup locally, usually with:

```env
APP_DOMAIN=localhost
```

**Render the production-style env file**:

```bash
./scripts/render-env.sh prd .env.prd
```

**Start the production stack locally**:

```bash
./scripts/prod-build.sh prd
./scripts/prod-up.sh prd
```

**Test locally**:

```bash
BASE_URL=http://localhost ./scripts/prod-smoke-test.sh
```

**Open**:

```
http://localhost
http://localhost/login
http://localhost/admin/scenarios
```

**Stop it**:

```bash
./scripts/prod-down.sh prd
```

# Real production deployment

**On the production server, create or edit**:

`env/prd.env`

**Use real production values**:

```env
APP_ENV=prd
APP_DOMAIN=app.yourdomain.com

JWT_SECRET=replace-with-a-real-long-random-secret

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@yourdomain.com
DEV_ADMIN_PASSWORD=temporary-strong-password
DEV_ADMIN_NAME=Admin User

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=

CORS_ALLOW_ORIGINS=https://app.yourdomain.com
```

**Render the production env file**:

```bash
./scripts/render-env.sh prd .env.prd
```

**Build and start production**:

```bash
./scripts/prod-build.sh prd
./scripts/prod-up.sh prd
```

**Check logs**:

```bash
./scripts/prod-logs.sh prd
```

**Run smoke test**:

```bash
BASE_URL=https://app.yourdomain.com ./scripts/prod-smoke-test.sh
```

**Open**:

```
https://app.yourdomain.com
https://app.yourdomain.com/login
https://app.yourdomain.com/admin/scenarios
```

**After the initial admin login works, change production admin seeding to false**:

```env
DEV_SEED_ADMIN=false
```

**Then re-render and restart**:

```bash
./scripts/render-env.sh prd .env.prd
./scripts/prod-up.sh prd
```

# Production backup

**Back up SQLite**:

```bash
./scripts/prod-backup-sqlite.sh
```

**The backup will be written to**:

```bash
backups/
```

# Production update workflow

**After pulling new code on the production server**:

```bash
git pull
./scripts/render-env.sh prd .env.prd
./scripts/prod-build.sh prd
./scripts/prod-up.sh prd
BASE_URL=https://app.yourdomain.com ./scripts/prod-smoke-test.sh
```

# Production rollback basics

**If a deploy fails, inspect logs**:

```bash
./scripts/prod-logs.sh prd
```

**Stop the stack**:

```bash
./scripts/prod-down.sh prd
```

**Then restore the previous Git version and rebuild**:

```bash
git checkout <previous-good-commit>
./scripts/prod-build.sh prd
./scripts/prod-up.sh prd
```

**For database safety, take a backup before upgrades**:

```bash
./scripts/prod-backup-sqlite.sh
```

# Corrected summary
```
Development:
  ./scripts/dev-backend.sh
  ./scripts/dev-frontend.sh

Production simulation:
  ./scripts/prod-build.sh prd
  ./scripts/prod-up.sh prd
  BASE_URL=http://localhost ./scripts/prod-smoke-test.sh

Real production:
  edit env/prd.env
  ./scripts/render-env.sh prd .env.prd
  ./scripts/prod-build.sh prd
  ./scripts/prod-up.sh prd
  BASE_URL=https://app.yourdomain.com ./scripts/prod-smoke-test.sh
  ```