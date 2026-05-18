# Call It Cure It — Three-Environment Launch Sequence

This document describes the launch procedure for the current deployment architecture:

```text
Internet
  -> edge Caddy on 80/443
    -> environment-local Caddy
      -> frontend/backend
```

The three environments are:

| Environment | Domain | Branch | Server folder | Compose project | Container prefix |
|---|---|---|---|---|---|
| Development | `dev.callitcureit.com` | `development` | `/opt/CallItCureIt/development` | `callitcureit-dev` | `callitcureit-dev` |
| Test | `tst.callitcureit.com` | `test` | `/opt/CallItCureIt/test` | `callitcureit-tst` | `callitcureit-tst` |
| Production | `app.callitcureit.com` | `production` | `/opt/CallItCureIt/production` | `callitcureit-prd` | `callitcureit-prd` |

Only the edge Caddy publishes ports `80` and `443`. Each environment has its own backend, frontend, Caddy, env file, and SQLite volume.

---

## 1. Required files on all three branches

Before launching test and production, make sure the following files have been merged from `development` into `test` and `production`:

```text
Makefile
docker-compose.server.yml
deploy/Caddyfile
edge/Caddyfile
edge/docker-compose.edge.yml
backend/docker-entrypoint.sh
frontend/nginx.conf
scripts/init-server-env.sh
```

Verify Makefile targets exist:

```bash
cd /opt/CallItCureIt/development
grep -n "server-dev-build" Makefile
grep -n "server-test-build" Makefile
grep -n "server-prod-build" Makefile
grep -n "edge-up" Makefile

cd /opt/CallItCureIt/test
git checkout test
git pull
grep -n "server-test-build" Makefile

cd /opt/CallItCureIt/production
git checkout production
git pull
grep -n "server-prod-build" Makefile
```

If either `grep` returns nothing, that branch does not yet have the new deployment package.

---

## 2. Launch development

```bash
cd /opt/CallItCureIt/development

git checkout development
git pull

make server-init-env ENV=development
nano .env.development
```

Confirm:

```env
APP_ENV=development
APP_DOMAIN=dev.callitcureit.com
CONTAINER_PREFIX=callitcureit-dev
DEV_SEED_ADMIN=true
CORS_ALLOW_ORIGINS=https://dev.callitcureit.com
```

Build and launch:

```bash
make server-dev-build
make server-dev-up
make server-dev-backend-health
```

Expected containers:

```bash
docker ps --format "table {{.Names}}\t{{.Status}}" | grep callitcureit-dev
```

Expected:

```text
callitcureit-dev-backend
callitcureit-dev-frontend
callitcureit-dev-caddy
```

---

## 3. Launch test

```bash
cd /opt/CallItCureIt/test

git checkout test
git pull

make server-init-env ENV=test
nano .env.test
```

Confirm:

```env
APP_ENV=test
APP_DOMAIN=tst.callitcureit.com
CONTAINER_PREFIX=callitcureit-tst
DEV_SEED_ADMIN=true
CORS_ALLOW_ORIGINS=https://tst.callitcureit.com
```

Set a real test admin password:

```env
DEV_ADMIN_EMAIL=admin-tst@callitcureit.com
DEV_ADMIN_PASSWORD=<test-admin-password>
DEV_ADMIN_NAME=Test Admin
```

Build and launch:

```bash
make server-test-build
make server-test-up
make server-test-backend-health
```

Expected containers:

```bash
docker ps --format "table {{.Names}}\t{{.Status}}" | grep callitcureit-tst
```

Expected:

```text
callitcureit-tst-backend
callitcureit-tst-frontend
callitcureit-tst-caddy
```

---

## 4. Launch production

```bash
cd /opt/CallItCureIt/production

git checkout production
git pull

make server-init-env ENV=production
nano .env.production
```

Confirm:

```env
APP_ENV=production
APP_DOMAIN=app.callitcureit.com
CONTAINER_PREFIX=callitcureit-prd
DEV_SEED_ADMIN=true
CORS_ALLOW_ORIGINS=https://app.callitcureit.com
```

Set strong production values:

```env
JWT_SECRET=<strong-production-secret>
DEV_ADMIN_EMAIL=admin@callitcureit.com
DEV_ADMIN_PASSWORD=<temporary-strong-production-admin-password>
DEV_ADMIN_NAME=Production Admin
```

Build and launch:

```bash
make server-prod-build
make server-prod-up
make server-prod-backend-health
```

Expected containers:

```bash
docker ps --format "table {{.Names}}\t{{.Status}}" | grep callitcureit-prd
```

Expected:

```text
callitcureit-prd-backend
callitcureit-prd-frontend
callitcureit-prd-caddy
```

---

## 5. Confirm all three Docker networks exist

After all three stacks are up, the edge proxy can attach to all three app networks.

```bash
docker network ls | grep callitcureit
```

You should see:

```text
callitcureit-dev_default
callitcureit-tst_default
callitcureit-prd_default
```

If one is missing, that environment stack is not up yet.

---

## 6. Configure edge proxy for all three environments

Edit:

```bash
nano /opt/CallItCureIt/edge/docker-compose.edge.yml
```

Use:

```yaml
services:
  edge-caddy:
    image: caddy:2.8-alpine
    container_name: callitcureit-edge-caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - edge-caddy-data:/data
      - edge-caddy-config:/config
    networks:
      - callitcureit-dev_default
      - callitcureit-tst_default
      - callitcureit-prd_default

volumes:
  edge-caddy-data:
  edge-caddy-config:

networks:
  callitcureit-dev_default:
    external: true
  callitcureit-tst_default:
    external: true
  callitcureit-prd_default:
    external: true
```

Then edit:

```bash
nano /opt/CallItCureIt/edge/Caddyfile
```

Use:

```caddyfile
dev.callitcureit.com {
	encode gzip zstd
	reverse_proxy callitcureit-dev-caddy:80
}

tst.callitcureit.com {
	encode gzip zstd
	reverse_proxy callitcureit-tst-caddy:80
}

app.callitcureit.com {
	encode gzip zstd
	reverse_proxy callitcureit-prd-caddy:80
}
```

---

## 7. Recreate edge Caddy

```bash
cd /opt/CallItCureIt/edge

docker compose -f docker-compose.edge.yml up -d --force-recreate
```

Check:

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep edge
```

Expected:

```text
callitcureit-edge-caddy
```

---

## 8. Test all three domains

```bash
curl -i https://dev.callitcureit.com/api/v1/healthz
curl -i https://tst.callitcureit.com/api/v1/healthz
curl -i https://app.callitcureit.com/api/v1/healthz
```

Expected for all three:

```text
HTTP/2 200
```

Then:

```bash
curl -i https://dev.callitcureit.com/login
curl -i https://tst.callitcureit.com/login
curl -i https://app.callitcureit.com/login
```

Expected for all three:

```text
HTTP/2 200
```

Run smoke tests:

```bash
cd /opt/CallItCureIt/development
make server-dev-smoke
make server-dev-admin-test

cd /opt/CallItCureIt/test
make server-test-smoke
make server-test-admin-test

cd /opt/CallItCureIt/production
make server-prod-smoke
make server-prod-admin-test
```

---

## 9. If admin login fails

If the admin user already exists with an older password, reset it for that environment.

For test:

```bash
cd /opt/CallItCureIt/test
make server-test-reset-admin
make server-test-admin-test
```

For production, only do this during first launch if no real production admin has been used yet:

```bash
cd /opt/CallItCureIt/production
make server-prod-reset-admin
make server-prod-admin-test
```

After production admin login works, edit:

```bash
nano /opt/CallItCureIt/production/.env.production
```

Set:

```env
DEV_SEED_ADMIN=false
```

Then:

```bash
cd /opt/CallItCureIt/production
make server-prod-up
make server-prod-smoke
make server-prod-admin-test
```

---

## 10. Final expected container state

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

Expected:

```text
callitcureit-edge-caddy      Up ... 0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp

callitcureit-dev-caddy       Up ...
callitcureit-dev-frontend    Up ...
callitcureit-dev-backend     Up ... healthy

callitcureit-tst-caddy       Up ...
callitcureit-tst-frontend    Up ...
callitcureit-tst-backend     Up ... healthy

callitcureit-prd-caddy       Up ...
callitcureit-prd-frontend    Up ...
callitcureit-prd-backend     Up ... healthy
```

Only the edge Caddy should publish ports `80` and `443`.

---

## 11. Short exact launch sequence

Once branches are merged and env files are ready:

```bash
cd /opt/CallItCureIt/development
git pull
make server-dev-build
make server-dev-up

cd /opt/CallItCureIt/test
git pull
make server-test-build
make server-test-up

cd /opt/CallItCureIt/production
git pull
make server-prod-build
make server-prod-up

cd /opt/CallItCureIt/edge
docker compose -f docker-compose.edge.yml up -d --force-recreate

cd /opt/CallItCureIt/development
make server-dev-smoke
make server-dev-admin-test

cd /opt/CallItCureIt/test
make server-test-smoke
make server-test-admin-test

cd /opt/CallItCureIt/production
make server-prod-smoke
make server-prod-admin-test
```

# Keep the canonical edge/Caddyfile in the repository.
Copy/sync it to /opt/CallItCureIt/edge when deploying edge changes.

# Recommended rule

Treat this as the source of truth, depending on which branch you are deploying from:
```bash
/opt/CallItCureIt/development/edge/Caddyfile
/opt/CallItCureIt/test/edge/Caddyfile
/opt/CallItCureIt/production/edge/Caddyfile
```

**Treat this as a deployed copy:**

```bash
/opt/CallItCureIt/edge/Caddyfile
```

**So the procedure should be:**
```bash
cd /opt/CallItCureIt/development
git pull
make edge-sync
make edge-up
```

or, once the edge changes are promoted to production:

```bash
cd /opt/CallItCureIt/production
git pull
make edge-sync
make edge-up
```

# Recommended Makefile targets

**Use these:**

```yaml
EDGE_DIR := $(SERVER_ROOT)/edge

.PHONY: edge-init
edge-init:
	mkdir -p $(EDGE_DIR)
	cp -R edge/* $(EDGE_DIR)/
	@echo "Initialized edge proxy folder at $(EDGE_DIR)"

.PHONY: edge-sync
edge-sync:
	mkdir -p $(EDGE_DIR)
	rsync -av --delete edge/ $(EDGE_DIR)/
	@echo "Synced repository edge/ to $(EDGE_DIR)"

.PHONY: edge-diff
edge-diff:
	@echo "Comparing repository edge/ with live $(EDGE_DIR)"
	diff -ru edge $(EDGE_DIR) || true

.PHONY: edge-up
edge-up:
	cd $(EDGE_DIR) && docker compose -f docker-compose.edge.yml up -d

.PHONY: edge-restart
edge-restart:
	cd $(EDGE_DIR) && docker compose -f docker-compose.edge.yml up -d --force-recreate

.PHONY: edge-reload
edge-reload:
	docker exec callitcureit-edge-caddy caddy reload --config /etc/caddy/Caddyfile

.PHONY: edge-deploy
edge-deploy: edge-sync edge-up edge-reload
	@echo "Edge proxy deployed."
```

# Practical policy

While you are stabilizing the architecture, I would sync edge from development:

```bash
cd /opt/CallItCureIt/development
make edge-sync
make edge-up
```

Once the deployment architecture is promoted through the branches, sync edge from production:

```bash
cd /opt/CallItCureIt/production
make edge-sync
make edge-up
```

That gives you a clean audit trail:

```
Git commit -> git pull -> make edge-sync -> make edge-up
```

and avoids mystery differences between repo files and live server files.

**One exception**

The only time I would edit /opt/CallItCureIt/edge/Caddyfile directly is during emergency debugging. But after that, immediately copy the fix back into the repository and deploy it properly.