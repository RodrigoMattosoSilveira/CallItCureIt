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

Then go to:

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