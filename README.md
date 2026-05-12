# CallItCureIt
Software application to train law students and young lawyers on what objections there are, be able to recognize when an objectionable question is asked or statement is said, and give a reason as to why it is objectionable.

# Recommended daily workflow
We will use 4 terminals:
- **Terminal 1 - Top Left**:
- **Terminal 2 - Top Left**: 
- **Terminal 3 - Bottom Left**:
- **Terminal 4 - Bottom Right**: 

## First-time or after schema changes:

```bash
# Teminal 1
$ make db-reset
```

## Run backend:

```bash
# Teminal 2
make dev-backend
```

## Run frontend in another terminal:

```bash
# Teminal 4
make dev-frontend
```

## Run all non-e2e checks:
```bash
# Teminal 3
make check
```

## Run e2e after backend/frontend are already running:
```bash
# Teminal 3
make rontend-e2e
```

## Run backend curl smoke test:
```bash
# Teminal 3
make test-hearsay-flow
```

## Run everything except server startup:
```bash
# Teminal 3
make check-with-e2e
```

# Scenation Administration
Manual flow:

1. Open /admin/scenarios
2. Open a scenario
3. Edit metadata
4. Add a line
5. Edit that line
6. Add opportunity
7. Edit opportunity
8. Delete opportunity
9. Delete line
10. Publish
11. Preview as trainee

# Recommended dev workflow

## Reset DB:

```bash
make db-reset
```

## Create admin:

```bash
cd backend
DATABASE_PATH=data/app.db \
JWT_SECRET=dev-secret-change-me \
ADMIN_EMAIL=admin@example.com \
ADMIN_PASSWORD=admin123 \
go run ./cmd/create-admin
```

## Run backend:

```bash
DATABASE_PATH=data/app.db JWT_SECRET=dev-secret-change-me go run ./cmd/api
```

## Run frontend:

```bash
cd frontend
npm run dev
```

## Open:
### iPhone
Ensure that the frontend `.env` has:
   `VITE_API_BASE_URL=http://192.168.2.154:8080/api/v1`
1. Start backend from backend/
   `DATABASE_PATH=data/app.db JWT_SECRET=dev-secret-change-me go run ./cmd/api`

2. Start frontend from frontend/
  `npm run dev`

3. Open on iPhone:
   `http://192.168.2.154:5173/login`

4. Log in:
   `admin@example.com`
   `admin123`

5. After login, navigate to:
   `http://192.168.2.154:5173/admin/scenarios`

Note that localhost only works from the same machine running Vite. On the iPhone, localhost means the iPhone itself, so the browser storage and API calls do not line up.

### local dev machine
```bash
http://localhost:5173/admin/scenarios

```

## You should be redirected to:

```bash
/login?redirectTo=%2Fadmin%2Fscenarios
```

## Login with:

```bash
admin@example.com
admin123
```