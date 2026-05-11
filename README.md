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