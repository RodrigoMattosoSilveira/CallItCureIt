# Production startup flow

## For local production simulation:

```bash
./scripts/init-env.sh
./scripts/prod-build.sh
./scripts/prod-up.sh
BASE_URL=http://localhost ./scripts/prod-smoke-test.sh
```

## For real production:

`cp backend/.env.production.example .env.production`

### Edit:

`.env.production`

### Set at least:

```
APP_DOMAIN=app.yourdomain.com
JWT_SECRET=a-long-random-secret
CORS_ALLOW_ORIGINS=https://app.yourdomain.com


DEV_SEED_ADMIN=true
DEV_ADMIN_EMAIL=admin@yourdomain.com
DEV_ADMIN_PASSWORD=temporary-strong-password
DEV_ADMIN_NAME=Admin User
```

### Then:

```bash
./scripts/prod-build.sh .env.production
./scripts/prod-up.sh .env.production
BASE_URL=https://app.yourdomain.com ./scripts/prod-smoke-test.sh
```

###v After you verify login works, change:

`DEV_SEED_ADMIN=false`

and restart:

```bash
./scripts/prod-up.sh .env.production
```