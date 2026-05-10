# CallItCureIt
software to train law students and young lawyers on what objections there are, be able to recognize when an objectionable question is asked or statement is said, and give a reason as to why it is objectionable.

# Manrual Browser Set
## Start backend:

```bash
cd backend
DATABASE_PATH=data/app.db go run ./cmd/api
```

## Start frontend:

```bash
cd frontend
npm run dev
```

## To work on scenarios:

```bash
http://localhost:5173/scenarios
```

Expected flow:

1. Click the seeded hearsay scenario.
2. Click Start Training Session.
3. Browser navigates to /sessions/:sessionId/play.
4. Click Next Line.
5. One transcript line appears.
6. Continue clicking Next Line.
7. At the end, session becomes completed.

## To admin on scenarios:

```bash
http://localhost:5173/admin/scenarios
```

Expected flow:

1. Click New Scenario.
2. Create a draft scenario.
3. Add transcript lines.
4. Add an objection opportunity to a line.
5. Publish the scenario.
6. Go to /scenarios and confirm it appears.
7. Start a training session from the newly created scenario.