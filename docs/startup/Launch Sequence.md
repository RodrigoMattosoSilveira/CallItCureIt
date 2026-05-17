# Introduction
Below is the exact launch sequence I recommend for the current architecture:

Current server model:
```
Internet
  -> edge Caddy on 80/443
    -> environment Caddy
      -> frontend/backend
```
Use this order:
1. Launch development app stack
2. Launch edge proxy
3. Validate development
4. Launch test app stack
5. Restart edge proxy
6. Validate test
7. Launch production app stack
8. Restart edge proxy
9. Validate production

# 0. One-time prerequisites

Run these once on the server.

## Confirm DNS

All three should point to the Hetzner server:
```bash
dig dev.callitcureit.com
dig tst.callitcureit.com
dig app.callitcureit.com
```
**Confirm server folders**
```bash
ls -la /opt/CallItCureIt
```

**Expected:**
```bash
development/
test/
production/
edge/
```

If edge/ does not exist yet, it will be created from the repo by make edge-init.

# 1. Development
## 1.1 Launch
```bash
cd /opt/CallItCureIt/development
git checkout development
git pull
```
**Initialize or update env file:**
```bash
make server-init-env ENV=development
nano .env.development
```
**Confirm the important values:**
```bash
APP_ENV=development
APP_DOMAIN=dev.callitcureit.com
CONTAINER_PREFIX=callitcureit-dev
DEV_SEED_ADMIN=true
CORS_ALLOW_ORIGINS=https://dev.callitcureit.com
```
**Build and launch:**
```bash
make server-dev-build
make server-dev-up
```
**Check containers:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep callitcureit-dev
```
**Expected:**
```bash
callitcureit-dev-backend
callitcureit-dev-frontend
callitcureit-dev-caddy
```
**Check internal backend:**
```bash
docker exec -it callitcureit-dev-backend curl -i http://localhost:8080/api/v1/healthz
```
Expected:
```bash
HTTP/1.1 200 OK
```
## 1.2. Launch Edge Proxy

Run this after the development stack exists.
```bash
cd /opt/CallItCureIt/development
make edge-init
make edge-up
```
**Check edge container:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep edge
```
**Expected:**
```bash
callitcureit-edge-caddy
```
If you are only launching development for now, your live /opt/CallItCureIt/edge/docker-compose.edge.yml must only attach to the dev network. If it references test/prod networks before they exist, edge startup will fail.

**Dev-only edge Compose:**
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

volumes:
  edge-caddy-data:
  edge-caddy-config:

networks:
  callitcureit-dev_default:
    external: true
```
**Dev-only edge Caddyfile:**
```bash
dev.callitcureit.com {
	encode gzip zstd
	reverse_proxy callitcureit-dev-caddy:80
}
```
## 1.3. Validate Development
```bash
curl -i https://dev.callitcureit.com/
curl -i https://dev.callitcureit.com/login
curl -i https://dev.callitcureit.com/scenarios
curl -i https://dev.callitcureit.com/api/v1/healthz
```
**Expected:**
```bash
/                       -> 200
/login                  -> 200
/scenarios              -> 200
/api/v1/healthz         -> 200
```
**Run smoke tests:**
```bash
make server-dev-smoke
make server-dev-admin-test
```
**If admin login fails, reset only the dev admin user:**
```bash
ADMIN_EMAIL=$(grep '^DEV_ADMIN_EMAIL=' .env.development | cut -d '=' -f2-)

docker exec -it callitcureit-dev-backend sqlite3 /app/data/app.db \
  "DELETE FROM users WHERE email = '$ADMIN_EMAIL';"

docker restart callitcureit-dev-backend
```
**Then rerun:**
```bash
make server-dev-admin-test
```
# 2. Test
## 2.1 Launch Test

Do this after the infrastructure changes have been merged into the test branch.
```bash
cd /opt/CallItCureIt/test
git checkout test
git pull
```
**Initialize or update env file:**
```bash
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
**Build and launch:**
```bash
make server-test-build
make server-test-up
```
**Check containers:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep callitcureit-tst
```
**Expected:**
```bash
callitcureit-tst-backend
callitcureit-tst-frontend
callitcureit-tst-caddy
```
**Check internal backend:**
```bash
docker exec -it callitcureit-tst-backend curl -i http://localhost:8080/api/v1/healthz
```
## 2.2 Update Edge Proxy for Dev + Test

**Edit or sync /opt/CallItCureIt/edge/Caddyfile so it has:**
```bash
dev.callitcureit.com {
	encode gzip zstd
	reverse_proxy callitcureit-dev-caddy:80
}

tst.callitcureit.com {
	encode gzip zstd
	reverse_proxy callitcureit-tst-caddy:80
}
```

**Edit or sync /opt/CallItCureIt/edge/docker-compose.edge.yml so it has both networks:**
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

volumes:
  edge-caddy-data:
  edge-caddy-config:

networks:
  callitcureit-dev_default:
    external: true
  callitcureit-tst_default:
    external: true
```
Restart edge:
```bash
cd /opt/CallItCureIt/development
make edge-up
```
or directly:
```bash
cd /opt/CallItCureIt/edge
docker compose -f docker-compose.edge.yml up -d --force-recreate
```
## 2.3 Validate Test
```bash
curl -i https://tst.callitcureit.com/
curl -i https://tst.callitcureit.com/login
curl -i https://tst.callitcureit.com/scenarios
curl -i https://tst.callitcureit.com/api/v1/healthz
```
**Expected all 200.**

**Run:**
```bash
cd /opt/CallItCureIt/test
make server-test-smoke
make server-test-admin-test
```
# 3. Production
## 3.1 Launch Production

Do this after the infrastructure changes have been merged into the production branch.
```bash
cd /opt/CallItCureIt/production
git checkout production
git pull
```
**Initialize or update env file:**
```bash
make server-init-env ENV=production
nano .env.production
```
**Confirm:**
```env
APP_ENV=production
APP_DOMAIN=app.callitcureit.com
CONTAINER_PREFIX=callitcureit-prd
DEV_SEED_ADMIN=true
CORS_ALLOW_ORIGINS=https://app.callitcureit.com
```
**For production, make sure these are strong:**
```bash
JWT_SECRET=<strong secret>
DEV_ADMIN_PASSWORD=<strong temporary admin password>
```
**Build and launch:**
```bash
make server-prod-build
make server-prod-up
```
**Check containers:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep callitcureit-prd
```
Expected:
```bash
callitcureit-prd-backend
callitcureit-prd-frontend
callitcureit-prd-caddy
```
**Check internal backend:**
```bash
docker exec -it callitcureit-prd-backend curl -i http://localhost:8080/api/v1/healthz\
```
## 3.2 Update Edge Proxy for All Three

**Final edge Caddyfile:**
```bash
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
**Final edge Compose:**
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
**Restart edge:**
```bash
cd /opt/CallItCureIt/edge
docker compose -f docker-compose.edge.yml up -d --force-recreate
```
Or from a repo folder if your Makefile has it:
```bash
make edge-up
```
## 3.3 Validate Production
```bash
curl -i https://app.callitcureit.com/
curl -i https://app.callitcureit.com/login
curl -i https://app.callitcureit.com/scenarios
curl -i https://app.callitcureit.com/api/v1/healthz
```
**Expected all 200.**

**Run:**
```bash
cd /opt/CallItCureIt/production
make server-prod-smoke
make server-prod-admin-test
```
**After first successful production admin login, edit:**
```bash
nano .env.production
```
**Set:**
```env
DEV_SEED_ADMIN=false
```
**Then recreate production stack:**
```bash
make server-prod-up
make server-prod-smoke
make server-prod-admin-test
```
# 4. Practical Current Shortcut

**Since development is working now, your immediate launch sequence for only dev is:**
```bash
cd /opt/CallItCureIt/development
git pull
make server-dev-build
make server-dev-up

make edge-init
make edge-up

make server-dev-smoke
make server-dev-admin-test
```
If edge currently references test/prod networks that do not exist yet, use the dev-only edge files until test/prod are launched.

# 5. Practical Current Shortcut
Final Launch Order Summary
## 5.1 Development
```bash
cd /opt/CallItCureIt/development
git checkout development
git pull
make server-init-env ENV=development
nano .env.development
make server-dev-build
make server-dev-up
make edge-init
make edge-up
make server-dev-smoke
make server-dev-admin-test
```
## 5.2 Test
```bash
cd /opt/CallItCureIt/test
git checkout test
git pull
make server-init-env ENV=test
nano .env.test
make server-test-build
make server-test-up
```
**Then update edge to include test and run:**
```bash
cd /opt/CallItCureIt/edge
docker compose -f docker-compose.edge.yml up -d --force-recreate
```
**Then:**
```bash
cd /opt/CallItCureIt/test
make server-test-smoke
make server-test-admin-test
```
## 5.3 Production
```bash
cd /opt/CallItCureIt/production
git checkout production
git pull
make server-init-env ENV=production
nano .env.production
make server-prod-build
make server-prod-up
```
**Then update edge to include production and run:**
```bash
cd /opt/CallItCureIt/edge
docker compose -f docker-compose.edge.yml up -d --force-recreate
```
**Then:**
```bash
cd /opt/CallItCureIt/production
make server-prod-smoke
make server-prod-admin-test
```
Then set:
```bash
DEV_SEED_ADMIN=false
```
and restart production.