# 1. Local Software Development Workflow

Purpose:

Developers write and test code on their local machines.

Flow:
1. Create or select a GitHub Issue.
2. Create a feature branch, usually from development.
3. Check out the feature branch locally.
4. Implement the change.
5. Run local backend/frontend checks.
6. Run local smoke tests.
7. Commit and push.
8. Open PR back into the source branch, usually development.

**Makefile-first commands**:
```bash
make local-init-env
make local-db-init
make local-check
make local-backend
make local-frontend
make local-smoke
```

**For issue work**:
```bash
git checkout development
git pull
git checkout -b issue-123-short-description
```

This Git workflow can remain explicit command-line because it is normal developer Git usage, though you could add helpers later.

# 2. Git Branch and Promotion Workflow
**Feature work**
```
GitHub Issue
  ↓
feature branch from development
  ↓
local implementation/testing
  ↓
PR into development
  ↓
deployment to dev.callitcureit.com
```
**Promotion to test**
```
development branch validated on dev
  ↓
PR or merge development → test
  ↓
deployment to tst.callitcureit.com
  ↓
app users and software engineers validate
```

**Promotion to production**
```
test branch validated on tst
  ↓
PR or merge test → production
  ↓
deployment to app.callitcureit.com
  ↓
production users use the app
```

For now, explicitly exclude emergency hotfixes:

Critical hotfixes to test or production branches are out of scope for this version of the process.

# 3. Shared Hetzner Server Layout

**Recommended structure**:

```bash
/opt/CallItCureIt/
  reverse-proxy/
  development/
  test/
  production/
```

**Initial setup**:
```
sudo mkdir -p /opt/CallItCureIt
sudo chown -R deploy:deploy /opt/CallItCureIt
```

**Then clone**:
```bash
cd /opt/CallItCureIt

git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git development
git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git test
git clone git@github.com:RodrigoMattosoSilveira/CallItCureIt.git production
```

**Then checkout branches**:
```bash
cd /opt/CallItCureIt/development
git checkout development

cd /opt/CallItCureIt/test
git checkout test

cd /opt/CallItCureIt/production
git checkout production
```

Subsequently, when preparing the environment
```bash:
# When preparing development
git pull origin development

# When preparing test
git pull origin test

# When preparing production
git pull origin production
```

# 4. Server Development Environment
## Summary
| Purpose | Domain | Branch | Folder | Env File |
|---------|--------|--------|--------|------------------|
| Intgration | dev.callitcureit.com | development | /opt/CallItCureIt/development | .env.development |

## Environment file
`.env.development`
```env
# Recommended environment values
APP_ENV=development
APP_DOMAIN=dev.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=<development-secret>
JWT_ISSUER=call-it-cure-it-development
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin-dev@callitcureit.com
DEV_ADMIN_PASSWORD=<development-admin-password>
DEV_ADMIN_NAME=Development Admin

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://dev.callitcureit.com
```
## Deployment procedure
```bash
cd /opt/CallItCureIt/development
make server-dev-pull
make server-dev-build
make server-dev-up
make server-dev-smoke
```

Where the Makefile target should internally use:
```
branch: development
compose project: callitcureit-dev
env file: .env.development
domain: dev.callitcureit.com
```
## Validation checklist
```
[ ] dev.callitcureit.com resolves to Hetzner
[ ] development branch checked out
[ ] .env.development exists
[ ] make server-dev-build succeeds
[ ] make server-dev-up starts containers
[ ] make server-dev-smoke passes
[ ] admin login works
[ ] scenario list works
[ ] scenario edit page works
[ ] training flow works
```
# 5. Server Test Environment
## Summary
| Purpose | Domain | Branch | Folder | Env File |
|---------|--------|--------|--------|------------------|
| Testing | tst.callitcureit.com | test | /opt/CallItCureIt/test | .env.test |

## Recommended environment values
```
APP_ENV=test
APP_DOMAIN=tst.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=<test-secret>
JWT_ISSUER=call-it-cure-it-test
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin-tst@callitcureit.com
DEV_ADMIN_PASSWORD=<test-admin-password>
DEV_ADMIN_NAME=Test Admin

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://tst.callitcureit.com
```
## Deployment procedure
```bash
cd /opt/CallItCureIt/test
make server-test-pull
make server-test-build
make server-test-up
make server-test-smoke
```
Where the Makefile target should internally use:
```
branch: test
compose project: callitcureit-tst
env file: .env.test
domain: tst.callitcureit.com
```
## Validation checklist
```
[ ] tst.callitcureit.com resolves to Hetzner
[ ] test branch checked out
[ ] .env.test exists
[ ] make server-test-build succeeds
[ ] make server-test-up starts containers
[ ] make server-test-smoke passes
[ ] app users validate the release candidate
[ ] software engineers validate regression areas
[ ] admin flow works
[ ] training flow works
[ ] no blocking browser console errors
[ ] no blocking backend errors
```
# 6. Server Production Environment
## Summary
| Purpose | Domain | Branch | Folder | Env File |
|---------|--------|--------|--------|------------------|
| For use | app.callitcureit.com | production | /opt/CallItCureIt/production | .env.production |

## Recommended environment values
```bash
APP_ENV=production
APP_DOMAIN=app.callitcureit.com

PORT=8080
DATABASE_PATH=/app/data/app.db

JWT_SECRET=<production-long-random-secret>
JWT_ISSUER=call-it-cure-it
JWT_EXPIRATION_MINUTES=480

DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@callitcureit.com
DEV_ADMIN_PASSWORD=<temporary-production-admin-password>
DEV_ADMIN_NAME=Production Admin

LLM_COACHING_ENABLED=false
OPENAI_API_KEY=
OPENAI_MODEL=gpt-5.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT_SECONDS=20

CORS_ALLOW_ORIGINS=https://app.callitcureit.com
```

## After first successful production admin login:
```bash
DEV_SEED_ADMIN=false
```
## Deployment procedure
```bash
cd /opt/CallItCureIt/production
make server-prod-pull
make server-prod-build
make server-prod-up
make server-prod-smoke
make server-prod-admin-test
```

Where the Makefile target should internally use:
```
branch: production
compose project: callitcureit-prd
env file: .env.production
domain: app.callitcureit.com
```
## Validation checklist
```
[ ] app.callitcureit.com resolves to Hetzner
[ ] production branch checked out
[ ] .env.production exists
[ ] JWT_SECRET is strong
[ ] admin password is strong
[ ] make server-prod-build succeeds
[ ] make server-prod-up starts containers
[ ] make server-prod-smoke passes
[ ] TLS certificate includes app.callitcureit.com
[ ] admin login works
[ ] DEV_SEED_ADMIN changed to false after first login
[ ] backup script works
[ ] production logs are clean
```
# 7. DNS A Records

The Hetzner server should have three DNS A records pointing to the same public IP:
```bash
dev.callitcureit.com  A  <Hetzner IP>
tst.callitcureit.com  A  <Hetzner IP>
app.callitcureit.com  A  <Hetzner IP>
```

**Validation targets**:
```
make server-dev-dns-check
make server-test-dns-check
make server-prod-dns-check
```
**Expected**:
```
dev.callitcureit.com -> <Hetzner IP>
tst.callitcureit.com -> <Hetzner IP>
app.callitcureit.com -> <Hetzner IP>
```

Also check that incorrect AAAA records do not point elsewhere.

# 8. Recommended Makefile Target Naming

I suggest using consistent target names:
## Local
```bash
make local-init-env
make local-db-init
make local-check
make local-backend
make local-frontend
make local-smoke
make local-admin-test
```
## Server development
```bash
make server-dev-checkout
make server-dev-pull
make server-dev-build
make server-dev-up
make server-dev-down
make server-dev-logs
make server-dev-smoke
make server-dev-admin-test
make server-dev-backup
make server-dev-dns-check
make server-dev-cert-check
```
## Server test
```bash
make server-test-checkout
make server-test-pull
make server-test-build
make server-test-up
make server-test-down
make server-test-logs
make server-test-smoke
make server-test-admin-test
make server-test-backup
make server-test-dns-check
make server-test-cert-check
```
## Server production
```bash
make server-prod-checkout
make server-prod-pull
make server-prod-build
make server-prod-up
make server-prod-down
make server-prod-logs
make server-prod-smoke
make server-prod-admin-test
make server-prod-backup
make server-prod-dns-check
make server-prod-cert-check
```
# 9. Suggested Makefile Abstraction

Internally, the Makefile can use variables like this:
```bash
ENV ?= development

ifeq ($(ENV),development)
  BRANCH := development
  DOMAIN := dev.callitcureit.com
  ENV_FILE := .env.development
  COMPOSE_PROJECT := callitcureit-dev
endif

ifeq ($(ENV),test)
  BRANCH := test
  DOMAIN := tst.callitcureit.com
  ENV_FILE := .env.test
  COMPOSE_PROJECT := callitcureit-tst
endif

ifeq ($(ENV),production)
  BRANCH := production
  DOMAIN := app.callitcureit.com
  ENV_FILE := .env.production
  COMPOSE_PROJECT := callitcureit-prd
endif
```
**Then generic commands can work**:
```bash
make server-pull ENV=development
make server-build ENV=development
make server-up ENV=development
make server-smoke ENV=development

make server-pull ENV=test
make server-build ENV=test
make server-up ENV=test
make server-smoke ENV=test

make server-pull ENV=production
make server-build ENV=production
make server-up ENV=production
make server-smoke ENV=production
```

**And friendly aliases can call them**:

```bash
server-dev-up:
	$(MAKE) server-up ENV=development

server-test-up:
	$(MAKE) server-up ENV=test

server-prod-up:
	$(MAKE) server-up ENV=production
```

This avoids duplicating all deployment logic three times.

# 10. Promotion Procedure
## Development integration
1. Developer creates GitHub Issue.
2. Developer creates feature branch from development.
3. Developer implements locally.
4. Developer runs local checks.
5. Developer pushes feature branch.
6. Developer opens PR into development.
7. PR reviewed and merged.
8. Server development environment is updated from development branch.
9. Software engineers validate dev.callitcureit.com.

**Makefile server update**:
```bash
cd /opt/CallItCureIt/development
make server-dev-pull
make server-dev-build
make server-dev-up
make server-dev-smoke
```

## Promote development to test
1. Development environment passes engineering validation.
2. Open PR from development into test.
3. Review and merge.
4. Update test server folder.
5. App users and software engineers validate tst.callitcureit.com.

**Makefile server update**:
```bash
cd /opt/CallItCureIt/test
make server-test-pull
make server-test-build
make server-test-up
make server-test-smoke
```
## Promote test to production
1. Test environment accepted.
2. Open PR from test into production.
3. Review and merge.
4. Update production server folder.
5. Run production deployment.
6. Run smoke tests.
7. Verify admin and public flows.
8. Monitor logs.

**Makefile server update**:
```bash
cd /opt/CallItCureIt/production
make server-prod-pull
make server-prod-build
make server-prod-up
make server-prod-smoke
make server-prod-admin-test
```
# 11. Shared Reverse Proxy Recommendation

**Because all three environments live on one server, do not run three independent Caddy containers all trying to bind**:
```
80
443
```

**Recommended approach**:
```
One server-level Caddy reverse proxy
Three app stacks without public ports
```
**Routes**:
```
dev.callitcureit.com → development frontend/backend
tst.callitcureit.com → test frontend/backend
app.callitcureit.com → production frontend/backend
```

**Possible server-level Caddyfile**:
```bash
dev.callitcureit.com {
    encode gzip zstd

    handle_path /api/* {
        reverse_proxy callitcureit-dev-backend:8080
    }

    handle {
        reverse_proxy callitcureit-dev-frontend:80
    }
}

tst.callitcureit.com {
    encode gzip zstd

    handle_path /api/* {
        reverse_proxy callitcureit-tst-backend:8080
    }

    handle {
        reverse_proxy callitcureit-tst-frontend:80
    }
}

app.callitcureit.com {
    encode gzip zstd

    handle_path /api/* {
        reverse_proxy callitcureit-prd-backend:8080
    }

    handle {
        reverse_proxy callitcureit-prd-frontend:80
    }
}
```

This requires container/network naming to be consistent.

If you want a simpler immediate path, deploy only one environment at a time with the existing Caddy-in-compose model. But for three simultaneous environments, use a shared reverse proxy.

# 12. Database and Volume Isolation

Each environment must have its own SQLite data volume.

Do not share one database between dev, test, and production.

**Recommended Compose project names create separate volumes automatically**:
```
callitcureit-dev_backend-data
callitcureit-tst_backend-data
callitcureit-prd_backend-data
```
**Backups should be environment-specific**:
```bash
make server-dev-backup
make server-test-backup
make server-prod-backup
```
**Backup locations**:
```bash
/opt/CallItCureIt/development/backups/
/opt/CallItCureIt/test/backups/
/opt/CallItCureIt/production/backups/
```
# 13. Revised Final Checklist
## Local
```
[ ] GitHub Issue created
[ ] Feature branch created from development
[ ] Branch checked out locally
[ ] make local-init-env
[ ] make local-db-init
[ ] make local-check
[ ] make local-backend
[ ] make local-frontend
[ ] make local-smoke
[ ] Browser training flow works
[ ] Browser admin flow works
[ ] Work committed and pushed
[ ] PR opened into development
```
## Server Development
```
[ ] dev.callitcureit.com DNS points to Hetzner
[ ] /opt/CallItCureIt/development exists
[ ] Repo cloned in development folder
[ ] development branch checked out
[ ] .env.development exists
[ ] make server-dev-pull
[ ] make server-dev-build
[ ] make server-dev-up
[ ] make server-dev-smoke
[ ] Engineers validate dev.callitcureit.com
```
## Server Test
```
[ ] tst.callitcureit.com DNS points to Hetzner
[ ] /opt/CallItCureIt/test exists
[ ] Repo cloned in test folder
[ ] test branch checked out
[ ] .env.test exists
[ ] development merged into test
[ ] make server-test-pull
[ ] make server-test-build
[ ] make server-test-up
[ ] make server-test-smoke
[ ] Users and engineers validate tst.callitcureit.com
```
## Server Production
```
[ ] app.callitcureit.com DNS points to Hetzner
[ ] /opt/CallItCureIt/production exists
[ ] Repo cloned in production folder
[ ] production branch checked out
[ ] .env.production exists
[ ] test merged into production
[ ] make server-prod-pull
[ ] make server-prod-build
[ ] make server-prod-up
[ ] make server-prod-smoke
[ ] make server-prod-admin-test
[ ] Production users can use app.callitcureit.com
[ ] DEV_SEED_ADMIN=false after initial admin bootstrap
[ ] Production backup verified
```
# My Suggested Final Adjustment

Your plan is good. My only strong recommendation is:

Use one shared Caddy/reverse-proxy deployment for the server,
not one Caddy per environment.

Everything else in your model is solid:
- one server
- three DNS A records
- three branches
- three folders
- three environment files
- three Compose project names
- three separate SQLite volumes
- promotion by PR/merge between branches

That gives you a clean and professional path:
local feature branch
  → development branch
  → dev.callitcureit.com
  → test branch
  → tst.callitcureit.com
  → production branch
  → app.callitcureit.com